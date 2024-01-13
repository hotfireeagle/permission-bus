package permissionbus

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

const menuType = "menu"
const apiType = "api"

type config struct {
	spec     string
	name     string
	chidlren []*config
}

type permissionBus struct {
	configData *config
}

func Load(filePath string) (*permissionBus, error) {
	pb := new(permissionBus)
	conf := new(config)

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return pb, err
	}

	err = json.Unmarshal(fileContent, conf)
	if err != nil {
		return pb, err
	}

	err = check(conf)
	if err != nil {
		return pb, err
	}

	pb.configData = conf
	return pb, nil
}

// 检查配置数据的格式:
// 1、所有的同等类型数据中，不允许出现一样的name
// 2、type是否正确
// 格式如果正确的话，那么将会返回true；如果错误的话，那么将会返回false
func check(conf *config) error {
	var err error

	existApiNameMap := make(map[string]bool)
	existMenuNameMap := make(map[string]bool)

	var dfs func(c *config)
	dfs = func(c *config) {
		if err != nil {
			// 检查已经出结果了，没必要接着在检查下去
			return
		}

		if c.spec == menuType {
			curMenu := c.name
			if existMenuNameMap[curMenu] {
				err = errors.New(curMenu + " repeat")
				return
			}
			existMenuNameMap[curMenu] = true
		} else if c.spec == apiType {
			curApi := c.name
			if existApiNameMap[curApi] {
				err = errors.New(curApi + " repeat")
				return
			}
			existApiNameMap[curApi] = true
		} else {
			err = errors.New("unsupport spec: " + c.spec)
			return
		}

		child := c.chidlren

		for _, item := range child {
			dfs(item)
		}
	}

	dfs(conf)

	return err
}

func (p *permissionBus) GetMenuTree() {

}

func (p *permissionBus) GetApiTree() {

}
