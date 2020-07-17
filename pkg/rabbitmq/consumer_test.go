package rabbitmq_test

import (
	"github.secureserver.net/digital-crimes/hashserve/pkg/rabbitmq"
	"testing"
)

type testBuildFileName struct {
	name 		string // Name to identify what the test is testing
	input		string // Input URL string
	output      string // SHA of the URL
}

var testBuildFileNames = []testBuildFileName{
	{
		name:"buildFileName JPG",
		input:"https://i.imgur.com/TZhncw3.jpg",
		output:"16bffea825fa1e450645eb416e234001d0b2d7dd",
	},
	{
		name:"buildFileName JPEG",
		input:"https://i.imgur.com/KxJLJ8b.jpeg",
		output:"a8a045f27beaaa65d5442db49ec539a76a704236",
	},
	{
		name:"buildFileName PNG",
		input:"https://i.imgur.com/FVoCOmP.png",
		output:"9caf43e0229ec8ac2cefc3e41109376e96b2775e",
	},
	{
		name:"buildFileName GIF",
		input:"https://i.imgur.com/17djyaF.mp4",
		output:"36a3704bd4b4101bf72c722335bc34b28d4a6786",
	},
}

func TestBuildFileName(t *testing.T) {
	for _, tt := range testBuildFileNames {
		generatedFileName := rabbitmq.BuildFileName(tt.input)
		expectedFileName:= tt.output
		if expectedFileName != generatedFileName {
			t.Errorf("Test(%s): expected %s, got %s", tt.name, tt.output, generatedFileName)
		}
	}
}

type testDownloadFile struct {
	name 		string // Name to identify what the test is testing
	input		string // Input URL string
	expectedFileName string //Expected file name
}

var testDownloadFiles = []testDownloadFile{
	{
		name:"downloadFile and fileExists JPG",
		input:"https://i.imgur.com/TZhncw3.jpg",
		expectedFileName:"16bffea825fa1e450645eb416e234001d0b2d7dd",
	},
	{
		name:"downloadFile and fileExists JPEG",
		input:"https://i.imgur.com/KxJLJ8b.jpeg",
		expectedFileName:"a8a045f27beaaa65d5442db49ec539a76a704236",
	},
	{
		name:"downloadFile and fileExists PNG",
		input:"https://i.imgur.com/FVoCOmP.png",
		expectedFileName:"9caf43e0229ec8ac2cefc3e41109376e96b2775e",
	},
	{
		name:"downloadFile and fileExists GIF",
		input:"https://i.imgur.com/17djyaF.mp4",
		expectedFileName:"36a3704bd4b4101bf72c722335bc34b28d4a6786",
	},
}

func TestDownloadFileAndFileExists(t *testing.T) {
	for _, tt := range testDownloadFiles {
		generatedFileName := rabbitmq.BuildFileName(tt.input)
		path:="/tmp/pdna/"+generatedFileName
		err := rabbitmq.DownloadFile(tt.input,path)
		if err != nil {
			t.Errorf("Test(%s): failed during file download", tt.name)
		}
		path ="/tmp/pdna/"+ tt.expectedFileName
		output := rabbitmq.FileExists(path)
		if !output {
			t.Errorf("Test(%s): expected %s, is not found locally", tt.name, tt.expectedFileName)
		}
	}
}

type testGenerateMD5Hash struct {
	name 		string // Name to identify what the test is testing
	input		string // Input file path
	expectedMD5 string // Expected MD5 Hash String
}

var testGenerateMD5Hashes = []testGenerateMD5Hash{
	{
		name:"generateMD5Hash JPG",
		input:"/tmp/pdna/16bffea825fa1e450645eb416e234001d0b2d7dd",
		expectedMD5:"690297404cd5bdd9cc09fc04bbe3b26b",
	},
	{
		name:"generateMD5Hash JPEG",
		input:"/tmp/pdna/a8a045f27beaaa65d5442db49ec539a76a704236",
		expectedMD5:"b771537b25aa9238e990b20abde6d170",
	},
	{
		name:"generateMD5Hash PNG",
		input:"/tmp/pdna/9caf43e0229ec8ac2cefc3e41109376e96b2775e",
		expectedMD5:"6fd1196dc7ea0e538f9cad8ec09ebdb3",
	},
	{
		name:"generateMD5Hash GIF",
		input:"/tmp/pdna/36a3704bd4b4101bf72c722335bc34b28d4a6786",
		expectedMD5:"8117fc44cefd77648b5983ead71daa87",
	},
}

func TestGenerateMD5Hash(t *testing.T){
	for _, tt := range testGenerateMD5Hashes {
		generatedMD5,err:=rabbitmq.GenerateMD5Hash(tt.input)
		if err != nil {
			t.Errorf("Test(%s): failed during file download", tt.name)
		}
		if generatedMD5 != tt.expectedMD5{
			t.Errorf("Test(%s): expected %s, got %s", tt.name,tt.expectedMD5, generatedMD5)
		}
	}
}