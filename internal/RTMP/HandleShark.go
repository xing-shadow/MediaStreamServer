package RTMP

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
)

/*
	rtmp handleShark: https://www.cnblogs.com/wanggang123/p/7513812.html
	C1:
		random fill 1536bytes c1 // also fill the c1-128bytes-key
		time = time() // c1[0-3]
		version = [0x80, 0x00, 0x07, 0x02] // c1[4-7]
		schema = choose schema0 or schema1
		digest-data = calc_c1_digest(c1, schema)
		calc_c1_digest(c1, schema) {
    		get c1s1-joined from c1 by specified schema
    		digest-data = HMACsha256(c1s1-joined, FPKey, 30)
    		return digest-data;
		}
		copy digest-data to c1
	S1:
		random fill 1536bytes s1
		time = time() // c1[0-3]
		version = [0x04, 0x05, 0x00, 0x01] // s1[4-7]
		DH_compute_key(key = c1-key-data, pub_key=s1-key-data)
		get c1s1-joined by specified schema
		s1-digest-data = HMACsha256(c1s1-joined, FMSKey, 36)
		copy s1-digest-data and s1-key-data to s1.
	C2:
		temp-key = HMACsha256(s1-digest, FPKey, 62)
		c2-digest-data = HMACsha256(c2-random-data, temp-key, 32)
	S2:
		temp-key = HMACsha256(c1-digest, FMSKey, 68)
		s2-digest-data = HMACsha256(s2-random-data, temp-key, 32)
*/

var (
	hsClientFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'P', 'l', 'a', 'y', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsServerFullKey = []byte{
		'G', 'e', 'n', 'u', 'i', 'n', 'e', ' ', 'A', 'd', 'o', 'b', 'e', ' ',
		'F', 'l', 'a', 's', 'h', ' ', 'M', 'e', 'd', 'i', 'a', ' ',
		'S', 'e', 'r', 'v', 'e', 'r', ' ',
		'0', '0', '1',
		0xF0, 0xEE, 0xC2, 0x4A, 0x80, 0x68, 0xBE, 0xE8, 0x2E, 0x00, 0xD0, 0xD1,
		0x02, 0x9E, 0x7E, 0x57, 0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
		0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
	}
	hsClientPartialKey = hsClientFullKey[:30]
	hsServerPartialKey = hsServerFullKey[:36]
)

func parseC1(c1 []byte, peerKey []byte) (ok bool, digest []byte) {
	var pos int
	if pos = findDigest(c1, peerKey, 772); pos == -1 { // key + digest
		if pos = findDigest(c1, peerKey, 8); pos == -1 { //digest + key
			return
		}
	}
	ok = true
	digest = c1[pos : pos+32]
	return
}

func findDigest(data []byte, key []byte, base int) int {
	pos := calcDigestPos(data, base)
	digest := makeDigest(key, data, pos)
	if bytes.Compare(digest, data[pos:pos+32]) != 0 {
		return -1
	}
	return pos
}

func calcDigestPos(c1 []byte, base int) (pos int) {
	for i := 0; i < 4; i++ {
		pos += int(c1[base+i])
	}
	pos = (pos % 728) + base + 4
	return
}

func makeDigest(key []byte, src []byte, pos int) (dst []byte) {
	h := hmac.New(sha256.New, key)
	if pos <= 0 {
		h.Write(src)
	} else {
		h.Write(src[:pos])
		h.Write(src[pos+32:])
	}
	return h.Sum(nil)
}

func createS0S1(src []byte, time uint32, ver uint32, key []byte) {
	src[0] = 3
	s1 := src[1:]
	rand.Read(s1[8:])
	binary.BigEndian.PutUint32(s1[0:4], time)
	binary.BigEndian.PutUint32(s1[4:8], ver)
	pos := calcDigestPos(s1, 8)
	digest := makeDigest(key, s1, pos)
	copy(s1[pos:], digest)
}

func CreateS2(src []byte, c1Digest []byte) {
	key := makeDigest(hsServerFullKey, c1Digest, -1)
	rand.Read(src)
	pos := len(src) - 32
	digest := makeDigest(key, src, pos)
	copy(src[pos:], digest)
}
