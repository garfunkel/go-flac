package main

import (
	"os"
	"bytes"
	"strings"
	"log"
	"fmt"
	"errors"
	"encoding/binary"
	"github.com/garfunkel/go-bitbuffer"
)

const (
	FLACMarker = "fLaC"
)

type BlockType uint

const (
	StreamInfo BlockType = iota
	Padding
	Application
	SeekTable
	VorbisComment
	CueSheet
	Picture
	Reserved
	Invalid = 127
)

type PictureType uint

const (
	Other PictureType = iota
	FileIcon
	OtherFileIcon
	FrontCover
	BackCover
	LeafletPage
	Media
	LeadArtist
	Artist
	Conductor
	Band
	Composer
	Lyricist
	RecordingLocation
	DuringRecording
	DuringPerformance
	ScreenCapture
	Fish
	Illustration
	BandLogo
	PublisherLogo
)

type SeekPoint struct {
	Sample uint64
	ByteOffset uint64
	NumSamples uint16
}

type IFLACMetadataBlock interface {
	parse(*os.File) error
	isLast() bool
}

type FLACMetadataBlock struct {
	FLAC *FLAC
	Last bool
	Type BlockType
	DataLength uint32
}

type FLACMetadataBlockStreamInfo struct {
	FLACMetadataBlock
	MinBlockSize uint16
	MaxBlockSize uint16
	MinFrameSize uint32
	MaxFrameSize uint32
	SampleRate uint32
	Channels uint8
	BitsPerSample uint8
	NumSamples uint64
	UnencodedMD5 []byte
}

type FLACMetadataBlockPadding struct {
	FLACMetadataBlock
	NumBytes uint32
}

type FLACMetadataBlockApplication struct {
	FLACMetadataBlock
	AppId string
	AppData []byte
}

type FLACMetadataBlockSeekTable struct {
	FLACMetadataBlock
	SeekPoints []SeekPoint
}

type FLACMetadataBlockVorbisComment struct {
	FLACMetadataBlock
	VendorString string
	Comments map[string][]string
}

type FLACMetadataBlockCueSheet struct {
	FLACMetadataBlock
}

type FLACMetadataBlockPicture struct {
	FLACMetadataBlock
	Type PictureType
	MIMEType string
	Description string
	Width uint32
	Height uint32
	ColourDepth uint32
	NumColours uint32
	Picture []byte
}

type FLACMetadataBlockReserved struct {
	FLACMetadataBlock
}

type FLAC struct {
	buffer *bitbuffer.BitBuffer
	Marker string
	StreamInfo *FLACMetadataBlockStreamInfo
	MetadataBlocks []IFLACMetadataBlock
}

func (block *FLACMetadataBlockStreamInfo) parse(handle *os.File) (err error) {
	blockData := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(blockData)

	block.FLACMetadataBlock.FLAC.buffer.Feed(blockData)
	data, err := block.FLACMetadataBlock.FLAC.buffer.ReadUint64(16)
	block.MinBlockSize = uint16(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(16)
	block.MaxBlockSize = uint16(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(24)
	block.MinFrameSize = uint32(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(24)
	block.MaxFrameSize = uint32(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(20)
	block.SampleRate = uint32(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(3)
	block.Channels = uint8(data) + 1
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(5)
	block.BitsPerSample = uint8(data) + 1
	block.NumSamples, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(36)
	block.UnencodedMD5, err = block.FLACMetadataBlock.FLAC.buffer.Read(128)

	return
}

func (block *FLACMetadataBlockStreamInfo) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockPadding) parse(handle *os.File) (err error) {
	blockData := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(blockData)

	block.NumBytes = block.FLACMetadataBlock.DataLength

	return
}

func (block *FLACMetadataBlockPadding) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockApplication) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)
	block.AppId, err = buffer.ReadString(32)
	block.AppData, err = buffer.Read(uint64(block.FLACMetadataBlock.DataLength * 8 - 32))

	return
}

func (block *FLACMetadataBlockApplication) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockSeekTable) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)

	for index := 0; index < len(data) / 18; index++ {
		seekPoint := SeekPoint{}

		seekPoint.Sample, err = buffer.ReadUint64(64)
		seekPoint.ByteOffset, err = buffer.ReadUint64(64)

		var numSamples uint64

		numSamples, err = buffer.ReadUint64(16)
		seekPoint.NumSamples = uint16(numSamples)

		block.SeekPoints = append(block.SeekPoints, seekPoint)
	}

	return
}

func (block *FLACMetadataBlockSeekTable) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockVorbisComment) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	buffer := bitbuffer.NewBitBuffer(binary.LittleEndian)

	buffer.Feed(data)

	length, err := buffer.ReadUint64(32)
	block.VendorString, err = buffer.ReadString(length * 8)
	length, err = buffer.ReadUint64(32)
	var commentLength uint64
	var comment string

	block.Comments = make(map[string][]string)

	for commentIndex := 0; commentIndex < int(length); commentIndex++ {
		commentLength, err = buffer.ReadUint64(32)
		comment, err = buffer.ReadString(commentLength * 8)
		commentFields := strings.SplitN(comment, "=", 2)
		
		if len(commentFields) != 2 {
			err = errors.New("malformed vorbis comment")

			return
		}

		block.Comments[commentFields[0]] = append(block.Comments[commentFields[0]], commentFields[1])
	}
	
	return
}

func (block *FLACMetadataBlockVorbisComment) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockCueSheet) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	return
}

func (block *FLACMetadataBlockCueSheet) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockPicture) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)

	return
}

func (block *FLACMetadataBlockPicture) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockReserved) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	return
}

func (block *FLACMetadataBlockReserved) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (flac *FLAC) parseMetadataBlock(handle *os.File) (block IFLACMetadataBlock, err error) {
	blockHeaderData := make([]byte, 4)

	_, err = handle.Read(blockHeaderData)

	if err != nil {
		return
	}

	lastBlock := (blockHeaderData[0] >> 7) != 0
	blockType := BlockType(blockHeaderData[0] << 1 >> 1)
	var dataLength uint32

	err = binary.Read(bytes.NewBuffer(blockHeaderData), binary.BigEndian, &dataLength)

	if err != nil {
		return
	}

	dataLength = (dataLength << 8 >> 8)

	blockHeader := FLACMetadataBlock{
		FLAC: flac,
		Last: lastBlock,
		Type: blockType,
		DataLength: dataLength,
	}

	switch blockType {
		case StreamInfo:
			block = &FLACMetadataBlockStreamInfo{
				FLACMetadataBlock: blockHeader,
			}

		case Padding:
			block = &FLACMetadataBlockPadding{
				FLACMetadataBlock: blockHeader,
			}

		case Application:
			block = &FLACMetadataBlockApplication{
				FLACMetadataBlock: blockHeader,
			}

		case SeekTable:
			block = &FLACMetadataBlockSeekTable{
				FLACMetadataBlock: blockHeader,
			}

		case VorbisComment:
			block = &FLACMetadataBlockVorbisComment{
				FLACMetadataBlock: blockHeader,
			}

		case CueSheet:
			block = &FLACMetadataBlockCueSheet{
				FLACMetadataBlock: blockHeader,
			}

		case Picture:
			block = &FLACMetadataBlockPicture{
				FLACMetadataBlock: blockHeader,
			}

		case Invalid:
			err = errors.New("Invalid")

			return

		default:
			block = &FLACMetadataBlockReserved{
				FLACMetadataBlock: blockHeader,
			}
	}

	block.parse(handle)

	return
}

func (flac *FLAC) parseStreamInfo(handle *os.File) (err error) {
	streamInfo, err := flac.parseMetadataBlock(handle)

	flac.StreamInfo = streamInfo.(*FLACMetadataBlockStreamInfo)

	return
}

func (flac *FLAC) parseStream(handle *os.File) (err error) {
	marker := make([]byte, 4)

	_, err = handle.Read(marker)

	if err != nil {
		return
	}

	flac.Marker = string(marker)

	if flac.Marker != FLACMarker {
		err = errors.New("FLAC marker not found")

		return
	}

	err = flac.parseStreamInfo(handle)

	if err != nil {
		return
	}

	last := flac.StreamInfo.FLACMetadataBlock.Last
	var iBlock IFLACMetadataBlock

	for !last {
		iBlock, err = flac.parseMetadataBlock(handle)

		if err != nil {
			return
		}

		flac.MetadataBlocks = append(flac.MetadataBlocks, iBlock)

		last = iBlock.isLast()
		fmt.Printf("%+v\n", iBlock)
	}

	return
}

func Parse(path string) (flac *FLAC, err error) {
	handle, err := os.Open(path)

	if err != nil {
		return
	}

	flac = &FLAC{
		buffer: bitbuffer.NewBitBuffer(binary.BigEndian),
	}

	err = flac.parseStream(handle)

	if err != nil {
		return
	}

	return
}

func main() {
	flac, err := Parse("/home/simon/Downloads/tone24bit.flac")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(flac)
	fmt.Printf("%+v\n", flac.StreamInfo)
}





