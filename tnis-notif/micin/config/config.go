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
	DB      *DBConfig
	ES      *ESConfig
	EMAIL   *EmailConfig
}

type SettingConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	User     string
	Password string
	Database string
}

type ESConfig struct {
	Host              string
	IndiceTransaction string
	IndiceCustomer    string
	IndiceToken       string
}

type EmailConfig struct {
	EmailAPI    string
	FromAddress string
	FromName    string
}

func GetConfig() *Config {
	var key = []byte("single sign on worknplay for all")
	var setting_config SettingConfig
	var db_config DBConfig
	var es_config ESConfig
	var email_config EmailConfig

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

	category = ".mysql_host"
	db_config.Host, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".mysql_user"
	db_config.User, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".mysql_password"
	db_config.Password, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".mysql_database"
	db_config.Database, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)

	category = ".elastic_host"
	es_config.Host, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".elastic_indice_transaction"
	es_config.IndiceTransaction, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".elastic_indice_customer"
	es_config.IndiceCustomer, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)
	category = ".elastic_indice_token"
	es_config.IndiceToken, err = Decrypt(key, viper.GetString(env+category))
	ErrorDecrypt(category, err)

	category = ".email_api"
	email_config.EmailAPI = viper.GetString(env + category)
	category = ".email_from_address"
	email_config.FromAddress = viper.GetString(env + category)
	category = ".email_from_name"
	email_config.FromName = viper.GetString(env + category)

	return &Config{
		SETTING: &SettingConfig{
			Port: setting_config.Port,
		},
		DB: &DBConfig{
			Host:     db_config.Host,
			User:     db_config.User,
			Password: db_config.Password,
			Database: db_config.Database,
		},
		ES: &ESConfig{
			Host:              es_config.Host,
			IndiceTransaction: es_config.IndiceTransaction,
			IndiceCustomer:    es_config.IndiceCustomer,
			IndiceToken:       es_config.IndiceToken,
		},
		EMAIL: &EmailConfig{
			EmailAPI:    email_config.EmailAPI,
			FromName:    email_config.FromName,
			FromAddress: email_config.FromAddress,
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
