package conf

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Configuration struct {
	Conf struct {
		SignerNodes    []string `yaml:"signerNodes"`
		ValidatorNodes []string `yaml:"validatorNodes"`

		KeyName        string   `yaml:"keyName"`
		KeyPath        string   `yaml:"keyPath"`
		Token			string  `yaml:"token"`


		//Signernode permissionless settings
		IsPermissionless  bool	`yaml:"isPermissionless"`
		IsOneTimeKey	bool	`yaml:"isOneTimeKey"`
		IsGroupRandomGenerated bool `yaml:"isGrupoRandomGenerated"`
		N int `yaml:"n"`
		T int `yaml:"t"`
		Scheme string `yaml:"scheme"`

		SendSignatureToAlgorand bool `yaml:"sendSignatureToAlgorand"`

	}
}

func ParseConfigFile(filename string) (Configuration, error) {
	buf, err := ioutil.ReadFile(filename)
	conf := Configuration{}

	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(buf, &conf)

	return conf, err
}
