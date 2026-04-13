package codecs

import (
	"strconv"
	"test.com/project-common/encrypts"
	projectModel "test.com/project-project/pkg/model"
)

func EncryptInt64(id int64) string {
	code, _ := encrypts.EncryptInt64(id, projectModel.AESKey)
	return code
}

func DecryptInt64(code string) (int64, error) {
	plain, err := encrypts.Decrypt(code, projectModel.AESKey)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(plain, 10, 64)
}
