package gomp3

/*
#define GOMP3_IMPLEMENTATION

#include "gomp3.h"
#include <stdlib.h>
#include <stdio.h>

int decode(mp3dec_t *dec, mp3dec_frame_info_t *info, unsigned char *data, int *length, unsigned char *decoded, int *decoded_length) {
    int samples;
    short pcm[GOMP3_MAX_SAMPLES_PER_FRAME];
    samples = mp3dec_decode_frame(dec, data, *length, pcm, info);
    *decoded_length = samples * info->channels * 2;
    *length -= info->frame_bytes;
    unsigned char buffer[samples * info->channels * 2];
    memcpy(buffer, (unsigned char*)&(pcm), sizeof(short) * samples * info->channels);
    memcpy(decoded, buffer, sizeof(short) * samples * info->channels);
    return info->frame_bytes;
}
*/
import "C"
import (
	"bytes"
	"fmt"
	"io"
	"unsafe"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
	"github.com/xxjwxc/public/message"
)

const maxSamplesPerFrame = 1152 * 2

// decoder decode the mp3 stream by gomp3
type decoder struct {
	mp3dec     C.mp3dec_t
	PcmData    []byte  // pcm 数据
	SampleRate int     // 采样率
	Channels   int     // 通道数
	Kbps       int     // 比特率
	Layer      int     // 层
	Time       float64 // 音频时长（s）
}

// NewMp3 put the mp3 data to decode.
func NewMp3(mp3 []byte) (dec *decoder, err error) {
	dec = new(decoder)
	dec.mp3dec = C.mp3dec_t{}
	C.mp3dec_init(&dec.mp3dec)
	info := C.mp3dec_frame_info_t{}
	var length = C.int(len(mp3))
	for {
		var decoded = [maxSamplesPerFrame * 2]byte{}
		var decodedLength = C.int(0)
		frameSize := C.decode(&dec.mp3dec,
			&info, (*C.uchar)(unsafe.Pointer(&mp3[0])),
			&length, (*C.uchar)(unsafe.Pointer(&decoded[0])),
			&decodedLength)
		if int(frameSize) == 0 {
			break
		}
		dec.PcmData = append(dec.PcmData, decoded[:decodedLength]...)
		if int(frameSize) < len(mp3) {
			mp3 = mp3[int(frameSize):]
		}
		dec.SampleRate = int(info.hz)
		dec.Channels = int(info.channels)
		dec.Kbps = int(info.bitrate_kbps)
		dec.Layer = int(info.layer)
	}
	dec.Time = (float64(len(dec.PcmData)) / (2.0 * float64(dec.SampleRate))) // 音频大小
	return
}

/*
*
numchannel:1=单声道，2=多声道
*/
func (dec *decoder) ToWav(numchannel int) ([]byte, error) {
	if dec == nil {
		return nil, message.GetErrorMsg(message.ParameterInvalid).GetError()
	}

	longSampleRate := dec.SampleRate
	byteRate := 16 * dec.SampleRate * numchannel / 8
	totalAudioLen := len(dec.PcmData)
	totalDataLen := totalAudioLen + 36
	var header = make([]byte, 44)
	// RIFF/WAVE header
	header[0] = 'R'
	header[1] = 'I'
	header[2] = 'F'
	header[3] = 'F'
	header[4] = byte(totalDataLen & 0xff)
	header[5] = byte((totalDataLen >> 8) & 0xff)
	header[6] = byte((totalDataLen >> 16) & 0xff)
	header[7] = byte((totalDataLen >> 24) & 0xff)
	//WAVE
	header[8] = 'W'
	header[9] = 'A'
	header[10] = 'V'
	header[11] = 'E'
	// 'fmt ' chunk
	header[12] = 'f'
	header[13] = 'm'
	header[14] = 't'
	header[15] = ' '
	// 4 bytes: size of 'fmt ' chunk
	header[16] = 16
	header[17] = 0
	header[18] = 0
	header[19] = 0
	// format = 1
	header[20] = 1
	header[21] = 0
	header[22] = byte(numchannel)
	header[23] = 0
	header[24] = byte(longSampleRate & 0xff)
	header[25] = byte((longSampleRate >> 8) & 0xff)
	header[26] = byte((longSampleRate >> 16) & 0xff)
	header[27] = byte((longSampleRate >> 24) & 0xff)
	header[28] = byte(byteRate & 0xff)
	header[29] = byte((byteRate >> 8) & 0xff)
	header[30] = byte((byteRate >> 16) & 0xff)
	header[31] = byte((byteRate >> 24) & 0xff)
	// block align
	header[32] = byte(2 * 16 / 8)
	header[33] = 0
	// bits per sample
	header[34] = 16
	header[35] = 0
	//data
	header[36] = 'd'
	header[37] = 'a'
	header[38] = 't'
	header[39] = 'a'
	header[40] = byte(totalAudioLen & 0xff)
	header[41] = byte((totalAudioLen >> 8) & 0xff)
	header[42] = byte((totalAudioLen >> 16) & 0xff)
	header[43] = byte((totalAudioLen >> 24) & 0xff)
	header = append(header, dec.PcmData...)
	return header, nil
}

/*
*
dst:二进制字符串
numchannel:1=单声道，2=多声道
saplerate：采样率 8000/16000
*/
func PcmToWav(byteDst []byte, numchannel int, saplerate int) (resDst string) {
	longSampleRate := saplerate
	byteRate := 16 * saplerate * numchannel / 8
	totalAudioLen := len(byteDst)
	totalDataLen := totalAudioLen + 36
	var header = make([]byte, 44)
	// RIFF/WAVE header
	header[0] = 'R'
	header[1] = 'I'
	header[2] = 'F'
	header[3] = 'F'
	header[4] = byte(totalDataLen & 0xff)
	header[5] = byte((totalDataLen >> 8) & 0xff)
	header[6] = byte((totalDataLen >> 16) & 0xff)
	header[7] = byte((totalDataLen >> 24) & 0xff)
	//WAVE
	header[8] = 'W'
	header[9] = 'A'
	header[10] = 'V'
	header[11] = 'E'
	// 'fmt ' chunk
	header[12] = 'f'
	header[13] = 'm'
	header[14] = 't'
	header[15] = ' '
	// 4 bytes: size of 'fmt ' chunk
	header[16] = 16
	header[17] = 0
	header[18] = 0
	header[19] = 0
	// format = 1
	header[20] = 1
	header[21] = 0
	header[22] = byte(numchannel)
	header[23] = 0
	header[24] = byte(longSampleRate & 0xff)
	header[25] = byte((longSampleRate >> 8) & 0xff)
	header[26] = byte((longSampleRate >> 16) & 0xff)
	header[27] = byte((longSampleRate >> 24) & 0xff)
	header[28] = byte(byteRate & 0xff)
	header[29] = byte((byteRate >> 8) & 0xff)
	header[30] = byte((byteRate >> 16) & 0xff)
	header[31] = byte((byteRate >> 24) & 0xff)
	// block align
	header[32] = byte(2 * 16 / 8)
	header[33] = 0
	// bits per sample
	header[34] = 16
	header[35] = 0
	//data
	header[36] = 'd'
	header[37] = 'a'
	header[38] = 't'
	header[39] = 'a'
	header[40] = byte(totalAudioLen & 0xff)
	header[41] = byte((totalAudioLen >> 8) & 0xff)
	header[42] = byte((totalAudioLen >> 16) & 0xff)
	header[43] = byte((totalAudioLen >> 24) & 0xff)

	headerDst := string(header)
	resDst = headerDst + string(byteDst)
	return
}

func ConvertMP3ToWAV(mp3Data []byte, targetSampleRate beep.SampleRate) ([]byte, error) {
	// 将 MP3 数据加载到内存中
	mp3Reader := bytes.NewReader(mp3Data)
	readCloser := NewReadCloser(mp3Reader)
	// 解码 MP3 数据
	streamer, format, err := mp3.Decode(readCloser)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()

	streamer1 := beep.Resample(4, format.SampleRate, targetSampleRate, streamer)
	format.SampleRate = targetSampleRate
	buffer := &bytes.Buffer{}

	// 创建一个缓冲区来存储 WAV 数据
	ws := NewWriteSeeker(buffer)

	// 将音频数据写入 WAV 缓冲区
	err = wav.Encode(ws, streamer1, format)
	if err != nil {
		return nil, err
	}

	return ReadWriteSeeker(ws)
}

/*
*
dst:二进制字符串
numchannel:1=单声道，2=多声道
saplerate：采样率 8000/16000
outQuality: 压缩质量：0: highest; 9: lowest
*/
// func PcmToMp3(bytePcm []byte, numchannel int, saplerate, outQuality int) ([]byte, error) {
// 	// var data []byte
// 	// // 创建一个新的buffer并写入字节切片
// 	var buf bytes.Buffer

// 	wr, err := lame.NewWriter(&buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	wr.InSampleRate = saplerate   // input sample rate
// 	wr.InNumChannels = numchannel // number of channels: 1
// 	wr.OutMode = lame.MODE_STEREO // common, 2 channels
// 	wr.OutQuality = outQuality    // 0: highest; 9: lowest
// 	wr.OutSampleRate = saplerate  // output sample rate

// 	io.Copy(wr, bytes.NewReader(bytePcm))
// 	wr.Close()

// 	return buf.Bytes(), nil
// }

type readCloserWrapper struct {
	reader *bytes.Reader
}

func (rc *readCloserWrapper) Read(p []byte) (int, error) {
	return rc.reader.Read(p)
}

func (rc *readCloserWrapper) Close() error {
	// `bytes.Reader` 不需要真正的关闭操作，所以这里返回 nil
	return nil
}

func NewReadCloser(reader *bytes.Reader) io.ReadCloser {
	return &readCloserWrapper{reader: reader}
}

type writeSeekerWrapper struct {
	buffer *bytes.Buffer
	offset int64
}

func (ws *writeSeekerWrapper) Write(p []byte) (int, error) {
	// 写入数据到缓冲区
	n, err := ws.buffer.Write(p)
	ws.offset += int64(n)
	return n, err
}

func (ws *writeSeekerWrapper) Seek(offset int64, whence int) (int64, error) {
	// 根据 whence 设置偏移量
	switch whence {
	case io.SeekStart:
		ws.offset = offset
	case io.SeekCurrent:
		ws.offset += offset
	case io.SeekEnd:
		ws.offset = int64(ws.buffer.Len()) + offset
	default:
		return 0, fmt.Errorf("无效的 whence 值: %d", whence)
	}

	// 检查偏移量是否有效
	if ws.offset < 0 || ws.offset > int64(ws.buffer.Len()) {
		return 0, fmt.Errorf("偏移量超出范围")
	}

	return ws.offset, nil
}

func NewWriteSeeker(buffer *bytes.Buffer) io.WriteSeeker {
	return &writeSeekerWrapper{buffer: buffer, offset: 0}
}

func ReadWriteSeeker(ws io.WriteSeeker) ([]byte, error) {
	// 检查具体实现是否支持读取
	if wrapper, ok := ws.(*writeSeekerWrapper); ok {
		return wrapper.buffer.Bytes(), nil
	}
	return nil, fmt.Errorf("无法读取 io.WriteSeeker 的值")
}
