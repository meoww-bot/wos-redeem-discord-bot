package main

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
	"testing"
)

func TestSign(t *testing.T) {

	userID := "417561"
	time := "1736525774515"

	// sign: 1d06c4341904034bf8ce1f05a97cb54b
	// fid: 417561
	// time: 1736525774515

	data := url.Values{}
	data.Set("fid", userID)
	data.Set("time", time)

	signedData := appendSign(data)

	t.Log(signedData)
}

func TestCNSign(t *testing.T) {

	// https://github.com/Crosswind/wos-gift-code/issues/10
	// sign=2e5e9554fbcb20c92fa3fb56e59f489f&fid=141079955&time=1736167631701

	userID := "141079955"
	time := "1736167631701"

	data := url.Values{}
	data.Set("fid", userID)
	data.Set("time", time)

	signedData := CNappendSign(data)

	t.Log(signedData)
}

func CNappendSign(data url.Values) url.Values {
	SECRET := "Uiv#87#SPan.ECsp"
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(data.Get(k))
	}
	buf.WriteString(SECRET)

	hash := md5.Sum([]byte(buf.String()))
	sign := hex.EncodeToString(hash[:])

	data.Set("sign", sign)
	return data
}
