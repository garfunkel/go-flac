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
	suite.assert.Equal(suite.flac.Marker, "fLaC")
}

func (suite *FLACTestSuite) TestFLACMetadataBlockStreamInfo() {
	suite.assert.NotNil(suite.flac.StreamInfo)
	suite.assert.Equal(suite.flac.StreamInfo.FLACMetadataBlock.FLAC, suite.flac)
	suite.assert.False(suite.flac.StreamInfo.FLACMetadataBlock.Last)
	suite.assert.Equal(suite.flac.StreamInfo.FLACMetadataBlock.Type, StreamInfo)
	suite.assert.Equal(suite.flac.StreamInfo.FLACMetadataBlock.DataLength, 34)
	suite.assert.Equal(suite.flac.StreamInfo.MinBlockSize, 4096)
	suite.assert.Equal(suite.flac.StreamInfo.MaxBlockSize, 4096)
	suite.assert.Equal(suite.flac.StreamInfo.MinFrameSize, 7822)
	suite.assert.Equal(suite.flac.StreamInfo.MaxFrameSize, 17848)
	suite.assert.Equal(suite.flac.StreamInfo.SampleRate, 88200)
	suite.assert.Equal(suite.flac.StreamInfo.Channels, 2)
	suite.assert.Equal(suite.flac.StreamInfo.BitsPerSample, 24)
	suite.assert.Equal(suite.flac.StreamInfo.NumSamples, 793287)
	suite.assert.Equal(hex.EncodeToString(suite.flac.StreamInfo.UnencodedMD5),
		"29499b5e67ae77df6f8491329c4deb93")
}

func (suite *FLACTestSuite) TestFLACMetadataBlockSeekTable() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockSeekTable)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(block.FLACMetadataBlock.FLAC, suite.flac)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(block.FLACMetadataBlock.Type, SeekTable)
		suite.assert.Equal(block.FLACMetadataBlock.DataLength, 18)
		suite.assert.Equal(len(block.SeekPoints), 1)
		suite.assert.Equal(block.SeekPoints[0].Sample, 0)
		suite.assert.Equal(block.SeekPoints[0].ByteOffset, 0)
		suite.assert.Equal(block.SeekPoints[0].NumSamples, 4096)
	}

	suite.assert.Equal(testedBlocks, 1)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockApplication() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockApplication)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(block.FLACMetadataBlock.FLAC, suite.flac)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(block.FLACMetadataBlock.Type, Application)
		suite.assert.Equal(block.FLACMetadataBlock.DataLength, 8)
		suite.assert.Equal(block.AppId, "ATCH")
		suite.assert.Equal(string(block.AppData), "C@K3")
	}

	suite.assert.Equal(testedBlocks, 1)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockVorbisComment() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockVorbisComment)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(block.FLACMetadataBlock.FLAC, suite.flac)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(block.FLACMetadataBlock.Type, VorbisComment)
		suite.assert.Equal(block.FLACMetadataBlock.DataLength, 56)
		suite.assert.Equal(block.VendorString, "reference libFLAC 1.1.4 20070213")
		suite.assert.Equal(len(block.Comments), 1)

		value, ok := block.Comments["example"]

		suite.assert.True(ok)
		suite.assert.Equal(len(value), 1)
		suite.assert.Contains(value, "fish")
	}

	suite.assert.Equal(testedBlocks, 1)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockPicture() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockPicture)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(block.FLACMetadataBlock.FLAC, suite.flac)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(block.FLACMetadataBlock.Type, Picture)
		suite.assert.Equal(block.FLACMetadataBlock.DataLength, 1661438)
		suite.assert.Equal(block.Type, FrontCover)
		suite.assert.Equal(block.MIMEType, "image/jpeg")
		suite.assert.Equal(block.Description, "")
		suite.assert.Equal(block.Width, 2448)
		suite.assert.Equal(block.Height, 3264)
		suite.assert.Equal(block.ColourDepth, 24)
		suite.assert.Equal(block.NumColours, 0)
		suite.assert.Equal("c6f3cec420be726d74ca3ccfb7461f65", hex.EncodeToString(block.PictureMD5))
	}

	suite.assert.Equal(testedBlocks, 1)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockCueSheet() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockCueSheet)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(block.FLACMetadataBlock.FLAC, suite.flac)
		suite.assert.False(block.FLACMetadataBlock.Last)
		suite.assert.Equal(block.FLACMetadataBlock.Type, CueSheet)
		suite.assert.Equal(block.FLACMetadataBlock.DataLength, 576)
		suite.assert.Equal(len(block.MediaCatalogNumber), 128)

		for _, c := range block.MediaCatalogNumber {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.Equal(block.NumLeadInSamples, 0)
		suite.assert.False(block.IsCD)
		suite.assert.Equal(4, len(block.CueSheetTracks))

		suite.assert.Equal(0, block.CueSheetTracks[0].Offset)
		suite.assert.Equal(1, block.CueSheetTracks[0].Track)
		suite.assert.Equal(len(block.CueSheetTracks[0].ISRC), 12)

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
		suite.assert.Equal(len(block.CueSheetTracks[0].ISRC), 12)

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
		suite.assert.Equal(len(block.CueSheetTracks[2].ISRC), 12)

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
		suite.assert.Equal(len(block.CueSheetTracks[3].ISRC), 12)

		for _, c := range block.CueSheetTracks[3].ISRC {
			suite.assert.Equal('\x00', rune(c))
		}

		suite.assert.True(block.CueSheetTracks[3].IsAudio)
		suite.assert.False(block.CueSheetTracks[3].PreEmphasis)
		suite.assert.Equal(0, len(block.CueSheetTracks[3].CueSheetTrackIndices))
	}

	suite.assert.Equal(testedBlocks, 1)
}

func (suite *FLACTestSuite) TestFLACMetadataBlockPadding() {
	testedBlocks := 0

	for _, iBlock := range suite.flac.MetadataBlocks {
		block, ok := iBlock.(*FLACMetadataBlockPadding)

		if !ok {
			continue
		}

		testedBlocks++

		suite.assert.Equal(block.FLACMetadataBlock.FLAC, suite.flac)
		suite.assert.True(block.FLACMetadataBlock.Last)
		suite.assert.Equal(block.FLACMetadataBlock.Type, Padding)
		suite.assert.Equal(block.FLACMetadataBlock.DataLength, 7596)
		suite.assert.Equal(block.NumBytes, 7596)
	}

	suite.assert.Equal(testedBlocks, 1)
}

func TestFLACTestSuite(t *testing.T) {
	suite.Run(t, new(FLACTestSuite))
}
