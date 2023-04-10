package collection

var setValue = struct{}{}

// SSet 由不重复的字符串组成的集合。该集合非线程安全。
type SSet struct {
	items map[string]struct{}
}

// NewSSet 返回指定容量的Set集合。
func NewSSet(capacity int, elements ...string) *SSet {
	var items map[string]struct{}
	if capacity <= 0 {
		capacity = len(elements)
	}
	items = make(map[string]struct{}, capacity)
	for i := range elements {
		items[elements[i]] = setValue
	}
	return &SSet{
		items: items,
	}
}

// Add 往Set集合中添加元素。
func (set *SSet) Add(elements ...string) *SSet {
	for i := range elements {
		set.items[elements[i]] = setValue
	}
	return set
}

// Remove 移除元素
func (set *SSet) Remove(elements ...string) *SSet {
	for i := range elements {
		delete(set.items, elements[i])
	}
	return set
}

// Elements 返回当前Set集合中的所有元素。
func (set *SSet) Elements() []string {
	all := make([]string, 0, len(set.items))
	for k := range set.items {
		all = append(all, k)
	}
	return all
}

// Contains 返回是否包含当前元素的布尔值
func (set *SSet) Contains(element string) bool {
	_, ok := set.items[element]
	return ok
}

// Length 返回当前的元素个数
func (set *SSet) Length() int {
	return len(set.items)
}

// IsEmpty 返回当前集合是否为空
func (set *SSet) IsEmpty() bool {
	return len(set.items) <= 0
}
