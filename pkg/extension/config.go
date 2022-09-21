package extension

var p *params

type params struct {
	value map[string]interface{}
}

func NewParams(v map[string]interface{}) {
	p = &params{value: v}
}

func (p *params) getGroups() (result []string) {
	if i, ok := p.value["defaultUserGroups"]; ok {
		for _, v := range i.([]interface{}) {
			result = append(result, v.(string))
		}
	}
	return
}

// FillDefaultGroups fill default user groups to idToken
func FillDefaultGroups(source []string) (result []string) {
	return distinctSlice(append(source, p.getGroups()...))
}

func distinctSlice(source []string) (result []string) {
	tempMap := map[string]bool{}
	for _, t := range source {
		if ok := tempMap[t]; !ok {
			tempMap[t] = true
			result = append(result, t)
		}
	}
	return
}
