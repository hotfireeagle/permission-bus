package permissionbus

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

const menuType = "menu"
const apiType = "api"
const apiGroupType = "apiGroup"

type PermissionConfigItem struct {
	Spec     string                 `json:"spec"`
	Name     string                 `json:"name"`
	Children []PermissionConfigItem `json:"children"`
	Group    []string               `json:"-"`
}

type PermissionBus struct {
	configData []PermissionConfigItem
}

func Load(filePath string) (*PermissionBus, error) {
	pb := new(PermissionBus)
	conf := new([]PermissionConfigItem)

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return pb, err
	}

	err = json.Unmarshal(fileContent, conf)
	if err != nil {
		return pb, err
	}

	err = checkNameNoRepeat(*conf)
	if err != nil {
		return pb, err
	}

	err = checkApiHasNoChildren(*conf)
	if err != nil {
		return pb, err
	}

	err = checkApiGroupNotContainMenu(*conf)
	if err != nil {
		return pb, err
	}

	pb.configData = *conf
	return pb, nil
}

// 检查配置数据的格式: 不允许出现name一样的数据
// 格式如果正确的话，那么将会返回true；如果错误的话，那么将会返回false
func checkNameNoRepeat(confs []PermissionConfigItem) error {
	var err error
	nameMap := make(map[string]bool)

	var dfs func(c PermissionConfigItem)
	dfs = func(c PermissionConfigItem) {
		if err != nil {
			// 检查已经出结果了，没必要接着在检查下去
			return
		}

		curName := c.Name
		if nameMap[curName] {
			err = errors.New(curName + " repeat")
			return
		}
		nameMap[curName] = true

		child := c.Children
		for _, item := range child {
			dfs(item)
		}
	}

	for _, conf := range confs {
		dfs(conf)
	}

	return err
}

// 检查配置数据：type为api时不允许出现children
func checkApiHasNoChildren(confs []PermissionConfigItem) error {
	var err error

	var dfs func(c PermissionConfigItem)
	dfs = func(c PermissionConfigItem) {
		if err != nil {
			return
		}

		if c.Spec == apiType {
			if len(c.Children) > 0 {
				err = errors.New("wrong format, we can't support the api permission has Children permission")
				return
			}
		}

		for _, confItem := range c.Children {
			dfs(confItem)
		}
	}

	for _, conf := range confs {
		dfs(conf)
	}

	return err
}

// 检查配置数据：type为apiGroup时，它的group里面不允许出现菜单
func checkApiGroupNotContainMenu(confs []PermissionConfigItem) error {
	// TODO:
	return nil
}

func (p *PermissionBus) GetMenuTree() []PermissionConfigItem {
	answer := make([]PermissionConfigItem, 0)

	var dfs func(c PermissionConfigItem) PermissionConfigItem
	dfs = func(c PermissionConfigItem) PermissionConfigItem {
		if c.Spec != menuType {
			return PermissionConfigItem{}
		}

		alternate := PermissionConfigItem{
			Spec: c.Spec,
			Name: c.Name,
		}

		childList := make([]PermissionConfigItem, 0)
		for _, item := range c.Children {
			copyItem := dfs(item)
			if copyItem.Spec != "" {
				childList = append(childList, copyItem)
			}
		}
		alternate.Children = childList

		return alternate
	}

	for _, c := range p.configData {
		copyItem := dfs(c)
		if copyItem.Spec != "" {
			answer = append(answer, copyItem)
		}
	}

	return answer
}

func (p *PermissionBus) GetApiTree() []PermissionConfigItem {
	return p.configData
}

// 只获取api级别，遇到apiGroup对其展开得到api，遇到menu过滤掉
func (p *PermissionBus) ExpandApiGroup(menuOrApiOrApiGroupList []string) []string {
	answer := make([]string, 0)
	flat := p.flatForExpandApiGroup()

	for _, name := range menuOrApiOrApiGroupList {
		item := flat[name]
		if item.Spec == apiType {
			answer = append(answer, name)
		} else if item.Spec == apiGroupType {
			for _, api := range item.Group {
				answer = append(answer, api)
			}
		}
	}

	return removeDuplicate(answer)
}

func (p *PermissionBus) GetMenuByLeaf() {

}

func (p *PermissionBus) flatForExpandApiGroup() map[string]PermissionConfigItem {
	answer := make(map[string]PermissionConfigItem)

	var dfs func(item PermissionConfigItem)
	dfs = func(conf PermissionConfigItem) {
		item := PermissionConfigItem{
			Name:  conf.Name,
			Spec:  conf.Spec,
			Group: conf.Group,
		}
		answer[conf.Name] = item
		for _, pci := range conf.Children {
			dfs(pci)
		}
	}

	for _, conf := range p.configData {
		// 由于目前只用于ExpandApiGroup，所以不希望copy children数据，children数据有点多（极端场景占内存），且用不上
		dfs(conf)
	}

	return answer
}

func removeDuplicate(cur []string) []string {
	m := make(map[string]bool)
	for _, n := range cur {
		m[n] = true
	}
	answer := make([]string, 0)
	for k := range m {
		answer = append(answer, k)
	}
	return answer
}
