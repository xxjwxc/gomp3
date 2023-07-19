# gomp3

[![Go Reference](https://pkg.go.dev/badge/github.com/xxjwxc/gomp3.svg)](https://pkg.go.dev/github.com/xxjwxc/gomp3) [![Builder](https://github.com/xxjwxc/gomp3/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/xxjwxc/gomp3/actions/workflows/ci.yaml) 
Decode mp3 base on <https://github.com/xxjwxc/gomp3>

## Installation

1. The first need Go installed (version 1.15+ is required), then you can use the below Go command to install gomp3.

``` bash
$ go get -u github.com/xxjwxc/gomp3
```

2. Import it in your code:

``` bash
import "github.com/xxjwxc/gomp3"
```

## Examples are here

<details>
  <summary>Example1: Decode the whole mp3 and play.</summary>

``` golang
package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/hajimehoshi/oto"
	"github.com/xxjwxc/gomp3"
)

func main() {
	var err error
	var file []byte
	if file, err = ioutil.ReadFile("./song.mp3"); err != nil {
		t.Error(err)
	}
	dec, err := gomp3.NewMp3(file)
	if err != nil {
		t.Error(err)
	}
	ioutil.WriteFile("song.pcm", dec.PcmData, 0644)// topcm

	data, _ := dec.ToWav(1)
	ioutil.WriteFile("song.wav", data, 0644)// towav

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
```
