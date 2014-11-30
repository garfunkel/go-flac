package main

import (
	"testing"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FLACTestSuite struct {
	suite.Suite
	flac *FLAC
	assert *assert.Assertions
}

func (suite *FLACTestSuite) SetupTest() {
	var err error

	suite.assert = assert.New(suite.T())
	suite.flac, err = Parse("sample.flac")

	suite.NoError(err)
	suite.assert.Equal("fLaC", suite.flac.Marker)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockStreamInfo() {
	suite.assert.NotNil(suite.flac.StreamInfo)
	suite.assert.Equal(suite.flac, suite.flac.StreamInfo.FLACMetadataBlock.FLAC)
	suite.assert.False(suite.flac.StreamInfo.FLACMetadataBlock.Last)
	suite.assert.Equal(StreamInfo, suite.flac.StreamInfo.FLACMetadataBlock.Type)
	suite.assert.Equal(34, suite.flac.StreamInfo.FLACMetadataBlock.DataLength)
	suite.assert.Equal(4096, suite.flac.StreamInfo.MinBlockSize)
	suite.assert.Equal(4096, suite.flac.StreamInfo.MaxBlockSize)
	suite.assert.Equal(7822, suite.flac.StreamInfo.MinFrameSize)
	suite.assert.Equal(17848, suite.flac.StreamInfo.MaxFrameSize)
	suite.assert.Equal(88200, suite.flac.StreamInfo.SampleRate)
	suite.assert.Equal(2, suite.flac.StreamInfo.Channels)
	suite.assert.Equal(24, suite.flac.StreamInfo.BitsPerSample)
	suite.assert.Equal(793287, suite.flac.StreamInfo.NumSamples)
	suite.assert.Equal("29499b5e67ae77df6f8491329c4deb93",
		hex.EncodeToString(suite.flac.StreamInfo.UnencodedMD5))
}

func (suite *FLACTestSuite) TestFLACMetadataBlockSeekTable() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockSeekTable)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(suite.flac, block.FLACMetadataBlock.FLAC)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(SeekTable, block.FLACMetadataBlock.Type)
		suite.assert.Equal(18, block.FLACMetadataBlock.DataLength)
		suite.assert.Equal(1, len(block.SeekPoints))
		suite.assert.Equal(0, block.SeekPoints[0].Sample)
		suite.assert.Equal(0, block.SeekPoints[0].ByteOffset)
		suite.assert.Equal(4096, block.SeekPoints[0].NumSamples)
	}

	suite.assert.Equal(1, testedBlocks)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockApplication() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockApplication)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(suite.flac, block.FLACMetadataBlock.FLAC)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(Application, block.FLACMetadataBlock.Type)
		suite.assert.Equal(8, block.FLACMetadataBlock.DataLength)
		suite.assert.Equal("ATCH", block.AppID)
		suite.assert.Equal("C@K3", string(block.AppData))
	}

	suite.assert.Equal(1, testedBlocks)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockVorbisComment() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockVorbisComment)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(suite.flac, block.FLACMetadataBlock.FLAC)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(VorbisComment, block.FLACMetadataBlock.Type)
		suite.assert.Equal(56, block.FLACMetadataBlock.DataLength)
		suite.assert.Equal("reference libFLAC 1.1.4 20070213", block.VendorString)
		suite.assert.Equal(1, len(block.Comments))

		value, ok := block.Comments["example"]

		suite.assert.True(ok)
		suite.assert.Equal(1, len(value))
		suite.assert.Contains(value, "fish")
	}

	suite.assert.Equal(1, testedBlocks)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockPicture() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockPicture)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(suite.flac, block.FLACMetadataBlock.FLAC)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(Picture, block.FLACMetadataBlock.Type)
		suite.assert.Equal(1661438, block.FLACMetadataBlock.DataLength)
		suite.assert.Equal(FrontCover, block.Type)
		suite.assert.Equal("image/jpeg", block.MIMEType)
		suite.assert.Equal("", block.Description)
		suite.assert.Equal(2448, block.Width)
		suite.assert.Equal(3264, block.Height)
		suite.assert.Equal(24, block.ColourDepth)
		suite.assert.Equal(0, block.NumColours)
		suite.assert.Equal("c6f3cec420be726d74ca3ccfb7461f65", hex.EncodeToString(block.PictureMD5))
	}

	suite.assert.Equal(1, testedBlocks)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockCueSheet() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockCueSheet)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(suite.flac, block.FLACMetadataBlock.FLAC)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(CueSheet, block.FLACMetadataBlock.Type)
		suite.assert.Equal(576, block.FLACMetadataBlock.DataLength)
		suite.assert.Equal(128, len(block.MediaCatalogNumber))

		for _, c := range block.MediaCatalogNumber {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.Equal(block.NumLeadInSamples, 0)
		suite.assert.False(block.IsCD)
		suite.assert.Equal(4, len(block.CueSheetTracks))

		suite.assert.Equal(0, block.CueSheetTracks[0].Offset)
		suite.assert.Equal(1, block.CueSheetTracks[0].Track)
		suite.assert.Equal(12, len(block.CueSheetTracks[0].ISRC))

		for _, c := range block.CueSheetTracks[0].ISRC {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.True(block.CueSheetTracks[0].IsAudio)
		suite.assert.False(block.CueSheetTracks[0].PreEmphasis)
		suite.assert.Equal(1, len(block.CueSheetTracks[0].CueSheetTrackIndices))
		suite.assert.Equal(0, block.CueSheetTracks[0].CueSheetTrackIndices[0].Offset)
		suite.assert.Equal(0, block.CueSheetTracks[0].CueSheetTrackIndices[0].IndexNumber)

		suite.assert.Equal(3528, block.CueSheetTracks[1].Offset)
		suite.assert.Equal(2, block.CueSheetTracks[1].Track)
		suite.assert.Equal(12, len(block.CueSheetTracks[0].ISRC))

		for _, c := range block.CueSheetTracks[0].ISRC {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.True(block.CueSheetTracks[1].IsAudio)
		suite.assert.False(block.CueSheetTracks[1].PreEmphasis)
		suite.assert.Equal(1, len(block.CueSheetTracks[1].CueSheetTrackIndices))
		suite.assert.Equal(0, block.CueSheetTracks[1].CueSheetTrackIndices[0].Offset)
		suite.assert.Equal(0, block.CueSheetTracks[1].CueSheetTrackIndices[0].IndexNumber)

		suite.assert.Equal(4704, block.CueSheetTracks[2].Offset)
		suite.assert.Equal(3, block.CueSheetTracks[2].Track)
		suite.assert.Equal(12, len(block.CueSheetTracks[2].ISRC))

		for _, c := range block.CueSheetTracks[2].ISRC {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.True(block.CueSheetTracks[2].IsAudio)
		suite.assert.False(block.CueSheetTracks[2].PreEmphasis)
		suite.assert.Equal(1, len(block.CueSheetTracks[2].CueSheetTrackIndices))
		suite.assert.Equal(0, block.CueSheetTracks[2].CueSheetTrackIndices[0].Offset)
		suite.assert.Equal(0, block.CueSheetTracks[2].CueSheetTrackIndices[0].IndexNumber)

		suite.assert.Equal(793287, block.CueSheetTracks[3].Offset)
		suite.assert.Equal(255, block.CueSheetTracks[3].Track)
		suite.assert.Equal(12, len(block.CueSheetTracks[3].ISRC))

		for _, c := range block.CueSheetTracks[3].ISRC {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.True(block.CueSheetTracks[3].IsAudio)
		suite.assert.False(block.CueSheetTracks[3].PreEmphasis)
		suite.assert.Equal(0, len(block.CueSheetTracks[3].CueSheetTrackIndices))
	}

	suite.assert.Equal(1, testedBlocks)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockPadding() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockPadding)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(suite.flac, block.FLACMetadataBlock.FLAC)
		suite.assert.True(block.FLACMetadataBlock.Last)
		suite.assert.Equal(Padding, block.FLACMetadataBlock.Type)
		suite.assert.Equal(7596, block.FLACMetadataBlock.DataLength)
		suite.assert.Equal(7596, block.NumBytes)
	}

	suite.assert.Equal(1, testedBlocks)
}

func TestFLACTestSuite(t *testing.T) {
	suite.Run(t, new(FLACTestSuite))
}
