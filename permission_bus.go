package permissionbus

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

const menuType = "menu"
const apiType = "api"

type PermissionConfigItem struct {
	Spec     string                 `json:"spec"`
	Name     string                 `json:"name"`
	Children []PermissionConfigItem `json:"children"`
}

type permissionBus struct {
	configData []PermissionConfigItem
}

func Load(filePath string) (*permissionBus, error) {
	pb := new(permissionBus)
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

	pb.configData = *conf
	return pb, nil
}

// 检查配置数据的格式:
// 1、所有的同等类型数据中，不允许出现一样的name
// 2、type是否正确
// 格式如果正确的话，那么将会返回true；如果错误的话，那么将会返回false
func checkNameNoRepeat(confs []PermissionConfigItem) error {
	var err error

	existApiNameMap := make(map[string]bool)
	existMenuNameMap := make(map[string]bool)

	var dfs func(c PermissionConfigItem)
	dfs = func(c PermissionConfigItem) {
		if err != nil {
			// 检查已经出结果了，没必要接着在检查下去
			return
		}

		if c.Spec == menuType {
			curMenu := c.Name
			if existMenuNameMap[curMenu] {
				err = errors.New(curMenu + " repeat")
				return
			}
			existMenuNameMap[curMenu] = true
		} else if c.Spec == apiType {
			curApi := c.Name
			if existApiNameMap[curApi] {
				err = errors.New(curApi + " repeat")
				return
			}
			existApiNameMap[curApi] = true
		} else {
			err = errors.New("unsupport Spec: " + c.Spec)
			return
		}

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

// 检查配置数据：api的type下面不允许出现children
// 因为在获取菜单树的时候，需要确定这块的规则
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

func (p *permissionBus) GetMenuTree() []PermissionConfigItem {
	answer := make([]PermissionConfigItem, 0)

	var dfs func(c PermissionConfigItem) PermissionConfigItem
	dfs = func(c PermissionConfigItem) PermissionConfigItem {
		if c.Spec == apiType {
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

func (p *permissionBus) GetApiTree() []PermissionConfigItem {
	return p.configData
}
