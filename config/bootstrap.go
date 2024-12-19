package config

import (
	"github.com/spf13/viper"
	"os"
)

// **嵌入文件只能在写embed指令的Go文件的同级目录或者子目录中
//
/*//go:embed *.yaml
var configs embed.FS*/

func init() {
	env := os.Getenv("ENV")
	vp := viper.New()

	// 根据环境变量 ENV 决定要读取的应用启动配置
	//configFileStream, err := configs.ReadFile("application." + env + ".yaml")
	//if err != nil {
	//	panic(err)
	//}

	if env == "" {
		env = "dev"
	}
	vp.SetConfigFile("./config/application." + env + ".yaml")
	err := vp.ReadInConfig()
	if err != nil {
		panic(err)
	}

	vp.UnmarshalKey("app", &App)
	vp.UnmarshalKey("database", &Database)

	vp.UnmarshalKey("redis", &Redis)
}
