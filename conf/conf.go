package conf

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Configuration struct {
	Conf struct {
		SignerNodes    []string `yaml:"signerNodes"`
		ValidatorNodes []string `yaml:"validatorNodes"`
	}
}

func ParseConfigFile(filename string) (Configuration,error) {
	buf, err := ioutil.ReadFile(filename)
	conf := Configuration{}

	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(buf, &conf)

	return conf,err
}
