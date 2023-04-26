package hashutils

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/bwmarrin/snowflake"
)

// HashSnowflake takes a snowflake as well as a salt value and returns a hash
func HashSnowflake(s string, salt []byte) (hash string, err error) {
	sid, err := snowflake.ParseString(s)
	if err != nil {
		return
	}

	idb := big.NewInt(sid.Int64() & int64(^uint(0)>>(64-48))).Bytes()
	comb := append(idb, salt...)
	hash = fmt.Sprintf("%x", sha256.Sum256(comb))

	return
}
