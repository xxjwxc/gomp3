package gomp3

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/hajimehoshi/oto"
)

func TestDecodeMp3(t *testing.T) {
	var err error
	var file []byte
	if file, err = ioutil.ReadFile("./song.mp3"); err != nil {
		t.Error(err)
	}
	dec, err := NewMp3(file)
	if err != nil {
		t.Error(err)
	}
	ioutil.WriteFile("song.pcm", dec.PcmData, 0644)

	data, _ := dec.ToWav(1)
	ioutil.WriteFile("song.wav", data, 0644)

	// play
	var context *oto.Context
	if context, err = oto.NewContext(dec.SampleRate, dec.Channels, 2, 1024); err != nil {
		log.Fatal(err)
	}

	var player = context.NewPlayer()
	player.Write(dec.PcmData)

	if err = player.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestDecodePcm(t *testing.T) {
	var err error
	var file []byte
	if file, err = ioutil.ReadFile("./song.pcm"); err != nil {
		t.Error(err)
	}

	out, _ := PcmToMp3(file, 1, 16000, 9)
	ioutil.WriteFile("out.mp3", out, 0644)

}
