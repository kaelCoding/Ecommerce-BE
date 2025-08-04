package utils

import (
  "golang.org/x/crypto/argon2" // Assuming you already have this package installed
  "encoding/base64"
  "errors"
  "fmt"
  "crypto/rand"
  "crypto/subtle"
  "strings"
  "strconv"
)

var (
  ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
  ErrIncompatibleVersion = errors.New("incompatible version of argon2")
  ErrEmptyPassword       = errors.New("password cannot be empty")
)

type HashParams struct {
  Memory      uint32
  Iterations  uint32
  Parallelism uint8
  SaltLength  uint32
  KeyLength   uint32
}

func GenerateFromPassword(password string, params *HashParams) (string, error) {
  if password == "" {
    return "", ErrEmptyPassword
  }
  
  salt, err := generateRandomBytes(params.SaltLength)
  if err != nil {
    return "", err
  }

  hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)
  b64Salt := base64.RawStdEncoding.EncodeToString(salt)
  b64Hash := base64.RawStdEncoding.EncodeToString(hash)
  encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Hash)

  return encodedHash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
  b := make([]byte, n)
  _, err := rand.Read(b)
  if err != nil {
    return nil, err
  }

  return b, nil
}

func ComparePasswordAndHash(password, encodedHash string) (bool, error) {
  p, salt, hash, err := decodeHash(encodedHash)
  if err != nil {
    return false, err
  }
  otherHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

  return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

func decodeHash(encodedHash string) (*HashParams, []byte, []byte, error) {
  values := strings.Split(encodedHash, "$")
  if len(values) != 6 {
    return nil, nil, nil, ErrInvalidHash
  }

  var version int
  _, err := fmt.Sscanf(values[2], "v=%d", &version)
  if err != nil {
    return nil, nil, nil, err
  }
  if version != argon2.Version {
    return nil, nil, nil, ErrIncompatibleVersion
  }

  p := &HashParams{}
  _, err = fmt.Sscanf(values[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
  if err != nil {
    return nil, nil, nil, err
  }

  salt, err := base64.RawStdEncoding.Strict().DecodeString(values[4])
  if err != nil {
    return nil, nil, nil, err
  }
  p.SaltLength = uint32(len(salt))

  hash, err := base64.RawStdEncoding.Strict().DecodeString(values[5])
  if err != nil {
    return nil, nil, nil, err
  }
  p.KeyLength = uint32(len(hash))

  return p, salt, hash, nil
}

func ParseUintOrError(s string) uint {
  parsed, err := strconv.ParseUint(s, 10, 32) 
  if err != nil {
    return 0 
  }
  return uint(parsed)
}