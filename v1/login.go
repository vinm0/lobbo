package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	util "io/ioutil"
	"log"
	"os"
	"time"

	"https://github.com/joho/godotenv"
)

//***** login *****

func autoLogin() client {
	loginFile := file{
		path: loginFilePath,
		name: loginFileName}

	loginKeys, err := loginFile.readFile()
	if err != nil {
		return client{}
	}
	for _, v := range loginKeys {
		if bytes.Equal(v[:lastLoginBlock], []byte("lastLogin: ")) {
			lastLogin, err := time.Parse(time.RFC822Z, string(v[lastLoginBlock:]))
			if err != nil {
				log.Println("Unable to determine last login: ", err)
				return client{}
			}
			if time.Now().After(lastLogin.Add(time.Hour * 24 * 30)) {
				return client{}
			}
		}
		if bytes.Equal(v[:usrnameBlock], []byte("username: ")) {
			usrname = v[usrnameBlock:]
		} else if bytes.Equal(v[:pwordBlock], []byte("password: ")) {
			pword = v[pwordBlock:]
		}
	} 
	decrypt(pword)
	c := login(usrname, pword)

	return c
}

func login(usrname, pword []byte) client {
	var c client 
	c.getClient(usrname, pword)
	return c
}


// ***** login file *****
func (f *file) readFile() ([][]byte, error) {
	var err error
	f.contents, err = util.ReadFile(f.path)
	if err != nil {
		s := ("login file not found")
		log.Println(s, err)
		return nil, err
	}
	if bytes.Contains(f.contents, loginBool) {
		loginKeys := make([][]byte, 0)
		r := bytes.NewReader(f.contents)
		bReader := bufio.NewReader(r)
		for item, err := bReader.ReadBytes('\n'); err != io.EOF; item, err = bReader.ReadBytes('\n') {
			loginKeys = append(loginKeys, item)
		}
		return loginKeys, nil
	}
	s := "Client not logged in or autologin unallowed"
	log.Println(s)
	return nil, errors.New(s)
}

func (f *file) writeToFile(w [][]byte) error {
	if w == nil {
		log.Println("No argument provided")
		return errors.New("nil array")
	}
	s := make([]byte, 0)
	for _, v := range w {
		s = append(s, v...) 
		s = append(s, []byte("\n")...)
	}
	err := util.WriteFile(loginFilePath, s, 0600)
	if err != nil {
		return errors.New("Cannot write to login file: ")
	}
	return nil
}

func (f *file) updateLastLogin() error {
	loginKeys, err := f.readFile()
	if err != nil {
		loginKeys = buildLogin
	}
	t := []byte(time.Now().String() + "\n")
	s := loginKeys[2][:lastLoginBlock]
	newTime := append(s, t...)

	loginKeys[2] = newTime

	f.writeToFile(loginKeys)
	return nil
}


// ***** encryption *****
var key = []byte(envVariable("LOGIN_SECRET_KEY"))

func envVariable(term string) string {
	// load .env file
	err := godotenv.Load("../.env")
	if err != nil {
	  log.Println("Error loading .env file")
	}
  
	return os.Getenv(term)
  }


func encrypt(pass []byte) ([]byte, error) {
	if pass == nil {
		return nil, errors.New("empty byte slice")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
        return nil, err
	}
    b := base64.StdEncoding.EncodeToString(pass)
    ciphertext := make([]byte, aes.BlockSize+len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
    return ciphertext, nil
}

func decrypt(text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
        return nil, err
    }
    if len(text) < aes.BlockSize {
        return nil, errors.New("ciphertext too short")
    }
    iv := text[:aes.BlockSize]
    text = text[aes.BlockSize:]
    cfb := cipher.NewCFBDecrypter(block, iv)
    cfb.XORKeyStream(text, text)
    data, err := base64.StdEncoding.DecodeString(string(text))
    if err != nil {
        return nil, err
    }
    return data, nil
}
