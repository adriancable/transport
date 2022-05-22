// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modifications copyright 2022 The Pion Authors, governed by
// the MIT license.

//go:build !amd64 && !ppc64 && !ppc64le && !arm64

package transport

import (
	"runtime"
	"unsafe"
)

const wordSize = int(unsafe.Sizeof(uintptr(0))) // nolint:gosec
const supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "ppc64" || runtime.GOARCH == "ppc64le" || runtime.GOARCH == "s390x" // nolint:gochecknoglobals

func isAligned(a *[]byte) bool {
	return uintptr(unsafe.Pointer(a)) % uintptr(wordSize) == 0
}

// xorBytes xors the bytes in a and b. The destination should have enough
// space, otherwise xorBytes will panic. Returns the number of bytes xor'd.
func XorBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	switch {
	case supportsUnaligned:
		fastXORBytes(dst, a, b, n)
	case isAligned(&dst) && isAligned(&a) && isAligned(&b):
		fastXORBytes(dst, a, b, n)
	default:
		safeXORBytes(dst, a, b, n)
	}
	return n
}

// fastXORBytes xors in bulk. It only works on architectures that
// support unaligned read/writes.
// n needs to be smaller or equal than the length of a and b.
func fastXORBytes(dst, a, b []byte, n int) {
	// Assert dst has enough space
	_ = dst[n-1]

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst)) // nolint:gosec
		aw := *(*[]uintptr)(unsafe.Pointer(&a)) // nolint:gosec
		bw := *(*[]uintptr)(unsafe.Pointer(&b)) // nolint:gosec
		for i := 0; i < w; i++ {
			dw[i] = aw[i] ^ bw[i]
		}
	}

	for i := (n - n%wordSize); i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}

// n needs to be smaller or equal than the length of a and b.
func safeXORBytes(dst, a, b []byte, n int) {
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}

// fastXORWords XORs multiples of 4 or 8 bytes (depending on architecture.)
// The arguments are assumed to be of equal length.
func fastXORWords(dst, a, b []byte) {
	dw := *(*[]uintptr)(unsafe.Pointer(&dst)) // nolint:gosec
	aw := *(*[]uintptr)(unsafe.Pointer(&a)) // nolint:gosec
	bw := *(*[]uintptr)(unsafe.Pointer(&b)) // nolint:gosec
	n := len(b) / wordSize
	for i := 0; i < n; i++ {
		dw[i] = aw[i] ^ bw[i]
	}
}

// fastXORWords XORs multiples of 4 or 8 bytes (depending on architecture.)
// The slice arguments a and b are assumed to be of equal length.
func XorWords(dst, a, b []byte) {
	if supportsUnaligned || (isAligned(&dst) && isAligned(&a) && isAligned(&b)) {
		fastXORWords(dst, a, b)
	} else {
		safeXORBytes(dst, a, b, len(b))
	}
}
