package crypt

type Xor struct {
	Key string
}

func NewXor(key string) *Xor {
	return &Xor{
		Key: key,
	}
}

func (xor *Xor) Encrypt(data []byte) ([]byte, error) {
	if len(xor.Key) == 0 {
		return data, nil
	}
	key := []byte(xor.Key)
	if len(key) > len(data) {
		key = key[:len(data)]
	}
	for i := range data {
		data[i] ^= key[i%len(key)]
	}
	return data, nil
}

func (xor *Xor) Decrypt(data []byte) ([]byte, error) {
	return xor.Encrypt(data)
}
