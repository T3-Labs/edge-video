package main

import (
	"sync"
)

// framePool é um pool global de buffers para frames
var framePool = sync.Pool{
	New: func() interface{} {
		// Aloca buffer de 2MB (suporta frames de diferentes resoluções)
		// cam1 RTMP: ~355KB, cam3 RTSP: ~184KB, cam2 RTSP: ~58KB
		buf := make([]byte, 2*1024*1024)
		return &buf
	},
}

// getFrameBuffer obtém um buffer do pool
func getFrameBuffer() *[]byte {
	return framePool.Get().(*[]byte)
}

// putFrameBuffer devolve um buffer ao pool
func putFrameBuffer(buf *[]byte) {
	if buf == nil {
		return
	}
	// Reset apenas o slice, não aloca memória nova
	*buf = (*buf)[:cap(*buf)]
	framePool.Put(buf)
}
