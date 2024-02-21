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
	Group    []string               `json:"group"`
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

	err = checkApiGroupMustHasGroupAndMustNoChildren(*conf)
	if err != nil {
		return pb, err
	}

	err = checkApiGroupNotContainMenuOrApiGroup(*conf)
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

// 检查配置数据：
// 1、type为api时不允许出现children
// 2、menu必须存在有效children
func checkApiHasNoChildren(confs []PermissionConfigItem) error {
	var err error

	checkMenuHasOtherSibling := func(list []PermissionConfigItem) bool {
		m := make(map[string]bool)
		for _, p := range list {
			m[p.Spec] = true
		}
		if m[menuType] {
			return len(m) > 1
		}
		return false
	}

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

		if c.Spec == menuType {
			if len(c.Children) == 0 {
				err = errors.New(c.Name + "wrong format, we can't support the menu permission has empty children")
				return
			}
		}

		if checkMenuHasOtherSibling(c.Children) {
			err = errors.New("menu just can has menu sibling in " + c.Children[0].Name)
			return
		}

		for _, confItem := range c.Children {
			dfs(confItem)
		}
	}

	if checkMenuHasOtherSibling(confs) {
		return errors.New("menu just can has menu sibling in " + confs[0].Name)
	}

	for _, conf := range confs {
		dfs(conf)
	}

	return err
}

// 检查配置数据：type为apiGroup时，它的group里面不允许出现菜单
func checkApiGroupNotContainMenuOrApiGroup(confs []PermissionConfigItem) error {
	menuOrApiGroupMap := make(map[string]bool)

	var dfsFindMenu func(p PermissionConfigItem)
	dfsFindMenu = func(p PermissionConfigItem) {
		if p.Spec == menuType || p.Spec == apiGroupType {
			menuOrApiGroupMap[p.Name] = true
		}

		for _, c := range p.Children {
			dfsFindMenu(c)
		}
	}
	for _, conf := range confs {
		dfsFindMenu(conf)
	}

	var err error
	var dfsCheck func(p PermissionConfigItem)
	dfsCheck = func(p PermissionConfigItem) {
		if err != nil {
			return
		}
		if p.Spec == apiGroupType {
			for _, name := range p.Group {
				if menuOrApiGroupMap[name] {
					err = errors.New(name + "出现在" + p.Name + ".Group中是不合法的")
					return
				}
			}
		}
		for _, c := range p.Children {
			dfsCheck(c)
		}
	}
	for _, co := range confs {
		dfsCheck(co)
	}

	return err
}

// 检查配置数据：
// 1、apiGroup类型的数据必须存在有效Group配置项
// 2、apiGroup类型的数据必须不存在Children配置
func checkApiGroupMustHasGroupAndMustNoChildren(confs []PermissionConfigItem) error {
	var err error

	var dfsCheck func(p PermissionConfigItem)
	dfsCheck = func(p PermissionConfigItem) {
		if err != nil {
			return
		}
		if p.Spec == apiGroupType {
			if len(p.Children) != 0 {
				err = errors.New(p.Name + "是apiGroup类型，不允许配置children")
				return
			}

			if len(p.Group) == 0 {
				err = errors.New(p.Name + "是apiGroup类型，必须存在Group")
			}
		}

		for _, c := range p.Children {
			dfsCheck(c)
		}
	}

	for _, c := range confs {
		dfsCheck(c)
	}

	return err
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
			// apiGroup的名称也应该加入进去，否则前端在显示选中效果的时候就不好处理
			// 同时由于下方存在去重的处理，也不需要担心同一个权限重复添加的情况
			answer = append(answer, name)
			for _, api := range item.Group {
				answer = append(answer, api)
			}
		}
	}

	return removeDuplicate(answer)
}

// 获取叶子结点的上级路径
func (p *PermissionBus) GetMenuByLeaf(leafs []string) []string {
	// 记录一个节点的子孙后代结点。第一层key是自己的name，第二层key是后代的name
	nodeChildrenMap := make(map[string]map[string]bool)
	var dfsFind func(p PermissionConfigItem)
	// 遍历权限树中的所有权限
	dfsFind = func(p PermissionConfigItem) {
		if p.Spec != menuType {
			return
		}
		name := p.Name
		childNameList := findChildren(p) // 自身的子孙节点
		for _, childName := range childNameList {
			if nodeChildrenMap[name] == nil {
				nodeChildrenMap[name] = make(map[string]bool)
			}
			nodeChildrenMap[name][childName] = true
		}
		for _, c := range p.Children {
			dfsFind(c)
		}
	}
	for _, n := range p.configData {
		dfsFind(n)
	}

	checkSelfHouDaiHasSelect := func(self map[string]bool) bool {
		for _, p := range leafs {
			if self[p] {
				return true
			}
		}
		return false
	}

	permissionNameAnswer := make([]string, 0)
	for permissNameKey, permissionChildObj := range nodeChildrenMap {
		if checkSelfHouDaiHasSelect(permissionChildObj) {
			permissionNameAnswer = append(permissionNameAnswer, permissNameKey)
		}
	}

	return permissionNameAnswer
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

// 找自己的后代结点。注：不仅仅只是子节点，还包括孙节点
func findChildren(p PermissionConfigItem) []string {
	answer := make([]string, 0)

	var dfs func(p PermissionConfigItem)
	dfs = func(p PermissionConfigItem) {
		for _, c := range p.Children {
			answer = append(answer, c.Name)
			dfs(c)
		}
	}
	dfs(p)

	return answer
}
