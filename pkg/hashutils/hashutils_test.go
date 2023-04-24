package hashutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	uid1 = "352002717285089280"
	uid2 = "531861558834495498"
	uid3 = "852813245815324672"
)

func TestHashUserID(t *testing.T) {
	{
		salt := []byte("pepper")

		hash1, err := HashUserID(uid1, salt)
		assert.Nil(t, err)

		hash2, err := HashUserID(uid1, salt)
		assert.Nil(t, err)

		assert.Equal(t, hash1, hash2)
	}

	{
		salt := []byte("salt")

		hash1, err := HashUserID(uid1, salt)
		assert.Nil(t, err)

		hash2, err := HashUserID(uid2, salt)
		assert.Nil(t, err)

		assert.NotEqual(t, hash1, hash2)
	}

	{
		salt1 := []byte("pepperv2")
		salt2 := []byte("saltv2")

		hash1, err := HashUserID(uid1, salt1)
		assert.Nil(t, err)

		hash2, err := HashUserID(uid1, salt2)
		assert.Nil(t, err)

		assert.NotEqual(t, hash1, hash2)
	}
}
