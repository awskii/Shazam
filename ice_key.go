// ICE encryption algorithm implementation.
package main

import "encoding/hex"

type IceKey struct {
	spBoxInitialised bool
	sMod             [4][4]int
	sXor             [4][4]int
	pBox             [32]uint
	keyrot           [16]int
	size             int
	rounds           int
	keySchedule      [16][3]int
	spBox            [4][1024]uint64
}

var (
	/* Modulo values for the S-boxes */
	ice_smod = [4][4]int{
		{333, 313, 505, 369},
		{379, 375, 319, 391},
		{361, 445, 451, 397},
		{397, 425, 395, 505}}
	/* XOR values for the S-boxes */
	ice_sxor = [4][4]int{
		{0x83, 0x85, 0x9b, 0xcd},
		{0xcc, 0xa7, 0xad, 0x41},
		{0x4b, 0x2e, 0xd4, 0x33},
		{0xea, 0xcb, 0x2e, 0x04}}
	/* Permutation values for the P-box */
	ice_pbox = [32]uint{
		0x00000001, 0x00000080, 0x00000400, 0x00002000,
		0x00080000, 0x00200000, 0x01000000, 0x40000000,
		0x00000008, 0x00000020, 0x00000100, 0x00004000,
		0x00010000, 0x00800000, 0x04000000, 0x20000000,
		0x00000004, 0x00000010, 0x00000200, 0x00008000,
		0x00020000, 0x00400000, 0x08000000, 0x10000000,
		0x00000002, 0x00000040, 0x00000800, 0x00001000,
		0x00040000, 0x00100000, 0x02000000, 0x80000000}
	/* The key rotation schedule */
	ice_keyrot = [16]int{
		0, 1, 2, 3, 2, 1, 3, 0,
		1, 3, 2, 0, 3, 1, 0, 2}
	keyschedule = [16][3]int{
		{767842, 728965, 791076},
		{617915, 557371, 577765},
		{976114, 799880, 978328},
		{664140, 141067, 432110},
		{508744, 716505, 976114},
		{297168, 300763, 690195},
		{744923, 974223, 956033},
		{404542, 545864, 148704},
		{767858, 243346, 939902},
		{1002018, 70183, 202391},
		{157635, 56046, 491236},
		{74308, 484035, 1017156},
		{461661, 519901, 315030},
		{519241, 192936, 148616},
		{257499, 297427, 919467},
		{937765, 215169, 973187},
	}
)

func NewIceKey() *IceKey {
	k := new(IceKey)
	k.Init(1)
	return k
}

// Initialize ICE key.
func (k *IceKey) Init(level int) {
	k.sMod = ice_smod
	k.sXor = ice_sxor
	k.pBox = ice_pbox
	k.keyrot = ice_keyrot
	if !k.spBoxInitialised {
		k.spBoxInit()
		k.spBoxInitialised = true
	}
	if level < 1 {
		k.size = 1
		k.rounds = 8
	} else {
		k.size = level
		k.rounds = level * 16
	}
	k.keySchedule = keyschedule
}

// Galois Field multiplication of a by b, modulo m.
// Just like arithmetic multiplication, except that additions and
// subtractions are replaced by XOR.
func (k *IceKey) gfMult(a, b, m int) int {
	num := 0
	for b != 0 {
		if (b & 1) != 0 {
			num ^= a
		}
		a <<= 1
		b >>= 1
		if a >= 256 {
			a ^= m
		}
	}
	return num
}

// Galois Field exponentiation.
// Raise the base to the power of 7, modulo m.
func (k *IceKey) gfExp7(b, m int) int64 {
	if b == 0 {
		return 0
	}
	x := k.gfMult(b, b, m)
	x = k.gfMult(b, x, m)
	x = k.gfMult(x, x, m)
	return int64(k.gfMult(b, x, m))
}

// Carry out the ICE 32-bit P-box permutation.
func (k *IceKey) perm32(x int64) uint64 {
	var res uint64 = 0
	for i := 0; x != 0; i++ {
		if (x & 1) != 0 {
			res |= uint64(k.pBox[i])
		}
		x >>= 1
	}
	return res
}

// Initialise the ICE S-boxes.
// This only has to be done once.
func (k *IceKey) spBoxInit() {
	for i := 0; i < 1024; i++ {
		col := (i >> 1) & 0xFF
		row := (i & 0x1) | ((i & 0x200) >> 8)
		x := k.gfExp7(col^k.sXor[0][row], k.sMod[0][row]) << 24
		k.spBox[0][i] = uint64(k.perm32(x))
		x = k.gfExp7(col^k.sXor[1][row], k.sMod[1][row]) << 16
		k.spBox[1][i] = uint64(k.perm32(x))
		x = k.gfExp7(col^k.sXor[2][row], k.sMod[2][row]) << 8
		k.spBox[2][i] = uint64(k.perm32(x))
		x = k.gfExp7(col^k.sXor[3][row], k.sMod[3][row])
		k.spBox[3][i] = uint64(k.perm32(x))
	}
}

// The single round ICE function.
func (k *IceKey) roundFunc(p uint64, i int, subkey [16][3]int) uint64 {
	tl := ((p >> 16) & 0x3ff) | (((p >> 14) | (p << 18)) & 0xffc00) // Left half expansion
	tr := (p & 0x3ff) | ((p << 2) & 0xffc00)                        // Left half expansion
	// Perform the salt permutation
	al := uint64(subkey[i][2]) & (tl ^ tr)
	ar := al ^ tr
	al ^= tl

	al ^= uint64(subkey[i][0])
	ar ^= uint64(subkey[i][1])

	// S-box lookup and permutation
	return k.spBox[0][al>>10] | k.spBox[1][al&0x3ff] | k.spBox[2][ar>>10] | k.spBox[3][ar&0x3ff]
}

// Encrypt a block of 8 bytes of data with the given ICE key.
func (k *IceKey) encrypt(ptext []byte, ctext *[]byte, idx int) {
	l := uint64(ptext[idx])<<24 | uint64(ptext[idx+1])<<16 | uint64(ptext[idx+2])<<8 | uint64(ptext[idx+3])
	r := uint64(ptext[idx+4])<<24 | uint64(ptext[idx+5])<<16 | uint64(ptext[idx+6])<<8 | uint64(ptext[idx+7])
	for i := 0; i < k.rounds; i += 2 {
		l ^= k.roundFunc(r, i, k.keySchedule)
		r ^= k.roundFunc(l, i+1, k.keySchedule)
	}
	for i := 0; i < 4; i++ {
		(*ctext)[idx+3-i] = byte(r & uint64(0xFF))
		(*ctext)[idx+7-i] = byte(l & uint64(0xFF))
		r >>= 8
		l >>= 8
	}
}

// Decrypt a block of 8 bytes of data with the given ICE key.
func (k *IceKey) decrypt(ciphertext []byte, plaintext *[]byte, idx int) {
	p1 := uint64(ciphertext[idx])<<24 | uint64(ciphertext[idx+1])<<16 | uint64(ciphertext[idx+2])<<8 | uint64(ciphertext[idx+3])
	p2 := uint64(ciphertext[idx+4])<<24 | uint64(ciphertext[idx+5])<<16 | uint64(ciphertext[idx+6])<<8 | uint64(ciphertext[idx+7])
	for i := k.rounds - 1; i > 0; i -= 2 {
		p1 ^= k.roundFunc(p2, i, k.keySchedule)
		p2 ^= k.roundFunc(p1, i-1, k.keySchedule)
	}
	for i := 0; i < 4; i++ {
		(*plaintext)[idx+3-i] = byte(p2 & uint64(0xff))
		(*plaintext)[idx+7-i] = byte(p1 & uint64(0xff))
		p2 >>= 8
		p1 >>= 8
	}
}

// Return the key size, in bytes.
func (k *IceKey) keySize() int { return k.size * 8 }

// Return the block size, in bytes.
func (k *IceKey) blockSize() int { return 8 }

func (k *IceKey) EncString(str string) string {
	length := (len(str)/8 + 1) * 8
	ptext := make([]byte, length)
	ctext := make([]byte, length)
	for k, c := range str {
		ptext[k] = byte(c)
	}
	for i := 0; i < length; i += 8 {
		k.encrypt(ptext, &ctext, i)
	}
	return "#0x" + hex.EncodeToString(ctext)
}

func (k *IceKey) EncBinary(data []byte) []byte {
	dataSize := len(data)
	length := (dataSize/8 + 1) * 8
	ptext := make([]byte, length)
	ctext := make([]byte, length)
	for i := 0; i < dataSize; i++ {
		ptext[i] = data[i]
	}
	for idx := 0; idx < length; idx += 8 {
		k.encrypt(ptext, &ctext, idx)
	}
	return ctext
}

func (k *IceKey) DecString(str string) []byte {
	str = str[3:]
	ptext := make([]byte, len(str)/2)
	buf, _ := hex.DecodeString(str)
	for i := 0; i < len(buf); i += 8 {
		k.decrypt(buf, &ptext, i)
	}
	return ptext
}

func (k *IceKey) Clear() {
	for i := 0; i < k.rounds; i++ {
		for j := 0; j < 3; j++ {
			k.keySchedule[i][j] = 0
		}
	}
}
