package config

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	SETTING *SettingConfig
	API     *APIConfig
}

type SettingConfig struct {
	Port string
}

type APIConfig struct {
	Auth        string
	Customer    string
	Transaction string
	Notif       string
}

func GetConfig() *Config {
	var key = []byte("single sign on worknplay for all")
	var setting_config SettingConfig
	var api_config APIConfig

	ex, err := os.Executable()
	if err != nil {
		fmt.Println("error config GetConfig os : ", err)
		return nil
	}

	myRealpath := path.Dir(ex)
	viper.SetConfigName("app")
	viper.AddConfigPath(myRealpath + "/config")
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println("error config GetConfig path : ", err)
		return nil
	}

	env := viper.GetString("env")

	category := ".port"
	setting_config.Port, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)

	category = ".api_auth"
	api_config.Auth, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".api_customer"
	api_config.Customer, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".api_transaction"
	api_config.Transaction, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".api_notif"
	api_config.Notif, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)

	return &Config{
		SETTING: &SettingConfig{
			Port: setting_config.Port,
		},
		API: &APIConfig{
			Auth:        api_config.Auth,
			Customer:    api_config.Customer,
			Transaction: api_config.Transaction,
			Notif:       api_config.Notif,
		},
	}
}

func Decrypt(key []byte, secure_mess string) (decoded_mess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(secure_mess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	decoded_mess = string(cipherText)
	return
}

func ErrorDecrypt(category string, err error) {
	if err != nil {
		fmt.Println("error config GetConfig decrypt %s : %s", category, err)
		return
	}
}
