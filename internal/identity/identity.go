package identity

const (
	Demander = "demander"
	Supplier = "supplier"
	Funder   = "funder"
)

var Labels = map[string]string{
	Demander: "项目方",
	Funder:   "资金方",
}

// ParentCategories 各身份可选的一级分类（行业）
var ParentCategories = map[string][]string{
	Demander: {"新能源项目", "企业类项目", "政信类项目", "其他类"},
	Funder:   {"租赁公司", "其他类"},
}

func Label(id string) string {
	if l, ok := Labels[id]; ok {
		return l
	}
	return id
}

func Valid(id string) bool {
	_, ok := Labels[id]
	return ok
}

func AllowedParents(id string) []string {
	if p, ok := ParentCategories[id]; ok {
		return p
	}
	return nil
}
