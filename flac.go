package main

import (
	"os"
	"bytes"
	"strings"
	"errors"
	"encoding/binary"
	"crypto/md5"
	"github.com/garfunkel/go-bitbuffer"
)

const (
	// FLACMarker is the standard FLAC identification string.
	FLACMarker = "fLaC"
)

// BlockType is the type used to identify the class of each metadata block.
type BlockType uint

// Enum indicating the type of each metadata block.
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

// PictureType is the type used to indicate picture format.
type PictureType uint

// Enum indicating the type of picture used in a metadata block.
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

// SeekPoint is a structure for storing the points at which a stream can be seeked.
type SeekPoint struct {
	Sample uint64
	ByteOffset uint64
	NumSamples uint16
}

// CueSheetTrackIndex is a structure for each cue index.
type CueSheetTrackIndex struct {
	Offset uint64
	IndexNumber uint8
}

// CueSheetTrack is a structure representing a cuesheet track.
type CueSheetTrack struct {
	Offset uint64
	Track uint8
	ISRC string
	IsAudio bool
	PreEmphasis bool
	CueSheetTrackIndices []CueSheetTrackIndex
}

// IFLACMetadataBlock is an interface for common behaviour of a metadata block.
type IFLACMetadataBlock interface {
	parse(*os.File) error
	isLast() bool
}

// FLACMetadataBlock sets out basic attributes for all metadata blocks.
type FLACMetadataBlock struct {
	FLAC *FLAC
	Last bool
	Type BlockType
	DataLength uint32
}

// FLACMetadataBlockStreamInfo sets out the structure for stream information.
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

// FLACMetadataBlockPadding represents padding metadata blocks.
type FLACMetadataBlockPadding struct {
	FLACMetadataBlock
	NumBytes uint32
}

// FLACMetadataBlockApplication represents application/binary metadata blocks.
type FLACMetadataBlockApplication struct {
	FLACMetadataBlock
	AppID string
	AppData []byte
}

// FLACMetadataBlockSeekTable represents the seek metadata block for a stream.
type FLACMetadataBlockSeekTable struct {
	FLACMetadataBlock
	SeekPoints []SeekPoint
}

// FLACMetadataBlockVorbisComment represents s tagging/vorbis comment metadata block.
type FLACMetadataBlockVorbisComment struct {
	FLACMetadataBlock
	VendorString string
	Comments map[string][]string
}

// FLACMetadataBlockCueSheet sets out the structure of a cuesheet metadata block.
type FLACMetadataBlockCueSheet struct {
	FLACMetadataBlock
	MediaCatalogNumber string
	NumLeadInSamples uint64
	IsCD bool
	CueSheetTracks []CueSheetTrack
}

// FLACMetadataBlockPicture sets out the structure used for a picture metadata block.
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
	PictureMD5 []byte
}

// FLACMetadataBlockReserved is an unused/reserved metadata block.
type FLACMetadataBlockReserved struct {
	FLACMetadataBlock
}

// FLAC is the primary structure for operations on FLAC files.
type FLAC struct {
	buffer *bitbuffer.BitBuffer
	Marker string
	StreamInfo *FLACMetadataBlockStreamInfo
	MetadataBlocks []IFLACMetadataBlock
}

func (block *FLACMetadataBlockStreamInfo) parse(handle *os.File) (err error) {
	blockData := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(blockData)

	if err != nil {
		return
	}

	block.FLACMetadataBlock.FLAC.buffer.Feed(blockData)
	data, err := block.FLACMetadataBlock.FLAC.buffer.ReadUint64(16)

	if err != nil {
		return
	}

	block.MinBlockSize = uint16(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(16)

	if err != nil {
		return
	}

	block.MaxBlockSize = uint16(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(24)

	if err != nil {
		return
	}

	block.MinFrameSize = uint32(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(24)

	if err != nil {
		return
	}

	block.MaxFrameSize = uint32(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(20)

	if err != nil {
		return
	}

	block.SampleRate = uint32(data)
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(3)

	if err != nil {
		return
	}

	block.Channels = uint8(data) + 1
	data, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(5)

	if err != nil {
		return
	}

	block.BitsPerSample = uint8(data) + 1
	block.NumSamples, err = block.FLACMetadataBlock.FLAC.buffer.ReadUint64(36)

	if err != nil {
		return
	}

	block.UnencodedMD5, err = block.FLACMetadataBlock.FLAC.buffer.Read(128)

	return
}

func (block *FLACMetadataBlockStreamInfo) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockPadding) parse(handle *os.File) (err error) {
	blockData := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(blockData)

	if err != nil {
		return
	}

	block.NumBytes = block.FLACMetadataBlock.DataLength

	return
}

func (block *FLACMetadataBlockPadding) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockApplication) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	if err != nil {
		return
	}

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)
	block.AppID, err = buffer.ReadString(32)

	if err != nil {
		return
	}

	block.AppData, err = buffer.Read(uint64(block.FLACMetadataBlock.DataLength * 8 - 32))

	return
}

func (block *FLACMetadataBlockApplication) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockSeekTable) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	if err != nil {
		return
	}

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)

	for index := 0; index < len(data) / 18; index++ {
		seekPoint := SeekPoint{}
		var numSamples uint64

		seekPoint.Sample, err = buffer.ReadUint64(64)

		if err != nil {
			return
		}

		seekPoint.ByteOffset, err = buffer.ReadUint64(64)

		if err != nil {
			return
		}

		numSamples, err = buffer.ReadUint64(16)

		if err != nil {
			return
		}

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

	if err != nil {
		return
	}

	buffer := bitbuffer.NewBitBuffer(binary.LittleEndian)

	buffer.Feed(data)

	length, err := buffer.ReadUint64(32)

	if err != nil {
		return
	}

	block.VendorString, err = buffer.ReadString(length * 8)

	if err != nil {
		return
	}

	length, err = buffer.ReadUint64(32)

	if err != nil {
		return
	}

	var commentLength uint64
	var comment string

	block.Comments = make(map[string][]string)

	for commentIndex := 0; commentIndex < int(length); commentIndex++ {
		commentLength, err = buffer.ReadUint64(32)

		if err != nil {
			return
		}

		comment, err = buffer.ReadString(commentLength * 8)

		if err != nil {
			return
		}

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

	if err != nil {
		return
	}

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)

	block.MediaCatalogNumber, err = buffer.ReadString(128 * 8)

	if err != nil {
		return
	}

	block.NumLeadInSamples, err = buffer.ReadUint64(64)
	
	if err != nil {
		return
	}

	isCD, err := buffer.ReadUint8(1)

	if err != nil {
		return
	}

	block.IsCD = isCD != 0

	_, err = buffer.Read(7 + 258 * 8)

	if err != nil {
		return
	}

	numTracks, err := buffer.ReadUint8(8)

	if err != nil {
		return
	}

	for trackIndex := uint8(0); trackIndex < numTracks; trackIndex++ {
		var flag uint8
		var numIndices uint8
		track := CueSheetTrack{}

		track.Offset, err = buffer.ReadUint64(64)

		if err != nil {
			return
		}

		track.Track, err = buffer.ReadUint8(8)

		if err != nil {
			return
		}

		track.ISRC, err = buffer.ReadString(12 * 8)

		if err != nil {
			return
		}

		flag, err = buffer.ReadUint8(1)

		if err != nil {
			return
		}

		track.IsAudio = flag == 0

		flag, err = buffer.ReadUint8(1)

		if err != nil {
			return
		}

		track.PreEmphasis = flag != 0

		_, err = buffer.Read(6 + 13 * 8)

		if err != nil {
			return
		}

		numIndices, err = buffer.ReadUint8(8)

		if err != nil {
			return
		}

		for indexIndex := uint8(0); indexIndex < numIndices; indexIndex++ {
			index := CueSheetTrackIndex{}

			index.Offset, err = buffer.ReadUint64(64)

			if err != nil {
				return
			}

			index.IndexNumber, err = buffer.ReadUint8(8)

			if err != nil {
				return
			}

			_, err = buffer.Read(3 * 8)

			if err != nil {
				return
			}

			track.CueSheetTrackIndices = append(track.CueSheetTrackIndices, index)
		}

		block.CueSheetTracks = append(block.CueSheetTracks, track)
	}

	return
}

func (block *FLACMetadataBlockCueSheet) isLast() bool {
	return block.FLACMetadataBlock.Last
}

func (block *FLACMetadataBlockPicture) parse(handle *os.File) (err error) {
	data := make([]byte, block.FLACMetadataBlock.DataLength)

	_, err = handle.Read(data)

	if err != nil {
		return
	}

	buffer := &block.FLACMetadataBlock.FLAC.buffer

	buffer.Feed(data)

	blockType, err := buffer.ReadUint32(32)

	if err != nil {
		return
	}

	block.Type = PictureType(blockType)

	mimeLength, err := buffer.ReadUint64(32)

	if err != nil {
		return
	}

	block.MIMEType, err = buffer.ReadString(mimeLength * 8)

	if err != nil {
		return
	}

	descLength, err := buffer.ReadUint64(32)

	if err != nil {
		return
	}

	block.Description, err = buffer.ReadString(descLength * 8)

	if err != nil {
		return
	}

	block.Width, err = buffer.ReadUint32(32)

	if err != nil {
		return
	}

	block.Height, err = buffer.ReadUint32(32)

	if err != nil {
		return
	}

	block.ColourDepth, err = buffer.ReadUint32(32)

	if err != nil {
		return
	}

	block.NumColours, err = buffer.ReadUint32(32)

	if err != nil {
		return
	}

	hasher := md5.New()
	picLength, err := buffer.ReadUint64(32)

	if err != nil {
		return
	}

	block.Picture, err = buffer.Read(picLength * 8)

	if err != nil {
		return
	}

	_, err = hasher.Write(block.Picture)

	if err != nil {
		return
	}

	block.PictureMD5 = hasher.Sum(nil)

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

	err = block.parse(handle)

	return
}

func (flac *FLAC) parseStreamInfo(handle *os.File) (err error) {
	streamInfo, err := flac.parseMetadataBlock(handle)

	if err != nil {
		return
	}

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
	}

	return
}

// Parse is the primary method for reading in a FLAC file and creating a handle.
func Parse(path string) (flac *FLAC, err error) {
	handle, err := os.Open(path)

	if err != nil {
		return
	}

	flac = &FLAC{
		buffer: bitbuffer.NewBitBuffer(binary.BigEndian),
	}

	err = flac.parseStream(handle)

	return
}
