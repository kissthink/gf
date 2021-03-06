// Copyright 2017 gf Author(https://github.com/gogf/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package gset provides kinds of concurrent-safe(alternative) sets.
//
// 并发安全集合.
package gset

import (
    "github.com/gogf/gf/g/internal/rwmutex"
    "github.com/gogf/gf/g/util/gconv"
    "strings"
)

type Set struct {
    mu *rwmutex.RWMutex
    m  map[interface{}]struct{}
}

// Create a set, which contains un-repeated items.
// The param <unsafe> used to specify whether using array with un-concurrent-safety,
// which is false in default, means concurrent-safe in default.
//
// 创建一个空的集合对象，参数unsafe用于指定是否用于非并发安全场景，默认为false，表示并发安全。
func New(unsafe...bool) *Set {
    return NewSet(unsafe...)
}

// See New.
//
// 同New.
func NewSet(unsafe...bool) *Set {
    return &Set{
        m  : make(map[interface{}]struct{}),
        mu : rwmutex.New(unsafe...),
    }
}

// Iterate the set by given callback <f>,
// if <f> returns true then continue iterating; or false to stop.
//
// 给定回调函数对原始内容进行遍历，回调函数返回true表示继续遍历，否则停止遍历。
func (set *Set) Iterator(f func (v interface{}) bool) *Set {
    set.mu.RLock()
    defer set.mu.RUnlock()
    for k, _ := range set.m {
        if !f(k) {
            break
        }
    }
    return set
}

// Add one or multiple items to the set.
//
// 添加元素项到集合中(支持多个).
func (set *Set) Add(item...interface{}) *Set {
    set.mu.Lock()
    for _, v := range item {
        set.m[v] = struct{}{}
    }
    set.mu.Unlock()
    return set
}

// Check whether the set contains <item>.
//
// 键是否存在.
func (set *Set) Contains(item interface{}) bool {
    set.mu.RLock()
    _, exists := set.m[item]
    set.mu.RUnlock()
    return exists
}

// Remove <item> from set.
//
// 删除元素项。
func (set *Set) Remove(item interface{}) *Set {
    set.mu.Lock()
    delete(set.m, item)
    set.mu.Unlock()
    return set
}

// Get size of the set.
//
// 获得集合大小。
func (set *Set) Size() int {
    set.mu.RLock()
    l := len(set.m)
    set.mu.RUnlock()
    return l
}

// Clear the set.
//
// 清空集合。
func (set *Set) Clear() *Set {
    set.mu.Lock()
    set.m = make(map[interface{}]struct{})
    set.mu.Unlock()
    return set
}

// Get the copy of items from set as slice.
//
// 获得集合元素项列表.
func (set *Set) Slice() []interface{} {
    set.mu.RLock()
    i   := 0
    ret := make([]interface{}, len(set.m))
    for item := range set.m {
        ret[i] = item
        i++
    }
    set.mu.RUnlock()
    return ret
}

// Join set items with a string.
//
// 使用glue字符串串连当前集合的元素项，构造成新的字符串返回。
func (set *Set) Join(glue string) string {
    return strings.Join(gconv.Strings(set.Slice()), ",")
}

// Return set items as a string, which are joined by char ','.
//
// 使用glue字符串串连当前集合的元素项，构造成新的字符串返回。
func (set *Set) String() string {
    return set.Join(",")
}

// Lock writing by callback function f.
//
// 使用自定义方法执行加锁修改操作。
func (set *Set) LockFunc(f func(m map[interface{}]struct{})) *Set {
    set.mu.Lock(true)
    defer set.mu.Unlock(true)
    f(set.m)
    return set
}

// Lock reading by callback function f.
//
// 使用自定义方法执行加锁读取操作。
func (set *Set) RLockFunc(f func(m map[interface{}]struct{})) *Set {
    set.mu.RLock(true)
    defer set.mu.RUnlock(true)
    f(set.m)
    return set
}

// Check whether the two sets equal.
//
// 判断两个集合是否相等.
func (set *Set) Equal(other *Set) bool {
    if set == other {
        return true
    }
    set.mu.RLock()
    defer set.mu.RUnlock()
    other.mu.RLock()
    defer other.mu.RUnlock()
    if len(set.m) != len(other.m) {
        return false
    }
    for key := range set.m {
        if _, ok := other.m[key]; !ok {
            return false
        }
    }
    return true
}

// Check whether the current set is sub-set of <other>.
//
// 判断当前集合是否为other集合的子集.
func (set *Set) IsSubsetOf(other *Set) bool {
    if set == other {
        return true
    }
    set.mu.RLock()
    defer set.mu.RUnlock()
    other.mu.RLock()
    defer other.mu.RUnlock()
    for key := range set.m {
        if _, ok := other.m[key]; !ok {
            return false
        }
    }
    return true
}

// Returns a new set which is the union of <set> and <other>.
// Which means, all the items in <newSet> is in <set> or in <other>.
//
// 并集, 返回新的集合：属于set或属于others的元素为元素的集合.
func (set *Set) Union(others ... *Set) (newSet *Set) {
    newSet = NewSet(true)
    set.mu.RLock()
    defer set.mu.RUnlock()
    for _, other := range others {
        if set != other {
            other.mu.RLock()
        }
        for k, v := range set.m {
            newSet.m[k] = v
        }
        if set != other {
            for k, v := range other.m {
                newSet.m[k] = v
            }
        }
        if set != other {
            other.mu.RUnlock()
        }
    }

    return
}

// Returns a new set which is the difference set from <set> to <other>.
// Which means, all the items in <newSet> is in <set> and not in <other>.
//
// 差集, 返回新的集合: 属于set且不属于others的元素为元素的集合.
func (set *Set) Diff(others...*Set) (newSet *Set) {
    newSet = NewSet(true)
    set.mu.RLock()
    defer set.mu.RUnlock()
    for _, other := range others {
        if set == other {
            continue
        }
        other.mu.RLock()
        for k, v := range set.m {
            if _, ok := other.m[k]; !ok {
                newSet.m[k] = v
            }
        }
        other.mu.RUnlock()
    }
    return
}

// Returns a new set which is the intersection from <set> to <other>.
// Which means, all the items in <newSet> is in <set> and also in <other>.
//
// 交集, 返回新的集合: 属于set且属于others的元素为元素的集合.
func (set *Set) Intersect(others...*Set) (newSet *Set) {
    newSet = NewSet(true)
    set.mu.RLock()
    defer set.mu.RUnlock()
    for _, other := range others {
        if set != other {
            other.mu.RLock()
        }
        for k, v := range set.m {
            if _, ok := other.m[k]; ok {
                newSet.m[k] = v
            }
        }
        if set != other {
            other.mu.RUnlock()
        }
    }
    return
}

// Returns a new set which is the complement from <set> to <full>.
// Which means, all the items in <newSet> is in <full> and not in <set>.
//
// 补集, 返回新的集合: (前提: set应当为full的子集)属于全集full不属于集合set的元素组成的集合.
// 如果给定的full集合不是set的全集时，返回full与set的差集.
func (set *Set) Complement(full *Set) (newSet *Set) {
    newSet = NewSet(true)
    set.mu.RLock()
    defer set.mu.RUnlock()
    if set != full {
        full.mu.RLock()
        defer full.mu.RUnlock()
    }
    for k, v := range full.m {
        if _, ok := set.m[k]; !ok {
            newSet.m[k] = v
        }
    }
    return
}