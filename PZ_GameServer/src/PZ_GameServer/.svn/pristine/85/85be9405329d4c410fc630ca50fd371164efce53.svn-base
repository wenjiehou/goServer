//
// 动态数组  Editor Kevin.Jia 2017-2-25
//

package arrayList

import "sync"

//  Example:
//
//		arr := arrayList.New()
//		n1 := ts{"n1"}
//		arr.Add(&n1)
//		arr.Add(&ts{"n2"})
//		arr.Add(&ts{"n3"})
//		fmt.Println(arr.Length())
//		arr.RemoveAt(0)
//		fmt.Println(arr.Length())
//		arr.Insert(arr.Count, &ts{"n4"})
//		t := (*arr.Index(1)).(*ts)
//		fmt.Println(t.Name)
//		fmt.Println((*arr.Index(arr.Count - 1)).(*ts).Name)

type ArrayList struct {
	mx    *sync.Mutex
	arr   []*interface{}
	Count int
}

func New() *ArrayList {
	return &ArrayList{
		arr:   make([]*interface{}, 0),
		mx:    new(sync.Mutex),
		Count: 0,
	}
}

//得到
func (a *ArrayList) Index(index int) *interface{} {
	if index < 0 || index >= a.Count {
		return nil
	} else {
		return a.arr[index]
	}
}

//添加
func (a *ArrayList) Add(obj interface{}) {
	a.mx.Lock()
	a.arr = append(a.arr, &obj)
	a.mx.Unlock()
	a.Count++
}

//删除
func (a *ArrayList) RemoveAt(index int) {
	if index >= 0 && index < a.Count {
		a.mx.Lock()
		a.arr = append(a.arr[:index], a.arr[index+1:]...)
		a.Count--
		a.mx.Unlock()
	}
}

//插入
func (a *ArrayList) Insert(index int, obj interface{}) {
	a.mx.Lock()
	a.arr = a.arr[0 : a.Count+1]
	copy(a.arr[index+1:], a.arr[index:])
	a.arr[index] = &obj
	a.mx.Unlock()
	a.Count++
}

//长度
func (a *ArrayList) Length() int {
	return len(a.arr)
}

//返回数组
func (a *ArrayList) Array() []*interface{} {
	return a.arr
}

//----------------------------------------------

//func New() *ArrayList {
//	return &ArrayList{
//		arr:   make([]interface{}, 0),
//		Count: 0,
//	}
//}

////得到
//func (a *ArrayList) Index(index int) interface{} {
//	if index < 0 || index >= a.Count {
//		return nil
//	} else {
//		return a.arr[index]
//	}
//}

////添加
//func (a *ArrayList) Add(obj interface{}) {
//	a.arr = append(a.arr, obj)
//	a.Count++
//}

////删除
//func (a *ArrayList) RemoveAt(index int) {
//	if index >= 0 || index < a.Count {
//		a.arr = append(a.arr[:index], a.arr[index+1:]...)
//		a.Count--
//	}
//}

////插入
//func (a *ArrayList) Insert(index int, obj interface{}) {
//	a.arr = a.arr[0 : a.Count+1]
//	copy(a.arr[index+1:], a.arr[index:])
//	a.arr[index] = obj
//	a.Count++
//}

////长度
//func (a *ArrayList) Length() int {
//	return len(a.arr)
//}
