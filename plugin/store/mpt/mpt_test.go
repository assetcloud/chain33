package mpt

import (
	"os"
	"testing"

	"fmt"
	"math/rand"
	"time"

	"github.com/stretchr/testify/assert"
	drivers "gitlab.33.cn/chain33/chain33/system/store"
	"gitlab.33.cn/chain33/chain33/types"
)

var store_cfg0 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test0", 100, false, false}
var store_cfg1 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test1", 100, false, false}
var store_cfg2 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test2", 100, false, false}
var store_cfg3 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test3", 100, false, false}

//var store_cfg4 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test4", 100,false, false}
//var store_cfg5 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test5", 100,false, false}
var store_cfg6 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test6", 100, false, false}
var store_cfg7 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test7", 100, false, false}
var store_cfg8 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test8", 100, false, false}
var store_cfg9 = &types.Store{"mpt_test", "leveldb", "/tmp/mpt_test9", 100, false, false}

const MaxKeylenth int = 64

func TestKvdbNewClose(t *testing.T) {
	os.RemoveAll(store_cfg0.DbPath) //删除已存在目录
	store := New(store_cfg0)
	assert.NotNil(t, store)

	store.Close()
}

func TestKvddbSetGet(t *testing.T) {
	os.RemoveAll(store_cfg1.DbPath)
	store := New(store_cfg1).(*Store)
	assert.NotNil(t, store)

	keys0 := [][]byte{[]byte("mk1"), []byte("mk2")}
	get0 := &types.StoreGet{drivers.EmptyRoot[:], keys0}
	values0 := store.Get(get0)
	mlog.Info("info", "info", values0)
	// Get exist key, result nil
	assert.Len(t, values0, 2)
	assert.Equal(t, []byte(nil), values0[0])
	assert.Equal(t, []byte(nil), values0[1])

	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{[]byte("k1"), []byte("v1")})
	kv = append(kv, &types.KeyValue{[]byte("k2"), []byte("v2")})
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	hash, err := store.Set(datas, true)
	assert.Nil(t, err)
	keys := [][]byte{[]byte("k1"), []byte("k2")}
	get1 := &types.StoreGet{hash, keys}

	values := store.Get(get1)
	assert.Len(t, values, 2)
	assert.Equal(t, []byte("v1"), values[0])
	assert.Equal(t, []byte("v2"), values[1])

	keys = [][]byte{[]byte("k1")}
	get2 := &types.StoreGet{hash, keys}
	values2 := store.Get(get2)
	assert.Len(t, values2, 1)
	assert.Equal(t, []byte("v1"), values2[0])

	get3 := &types.StoreGet{drivers.EmptyRoot[:], keys}
	values3 := store.Get(get3)
	assert.Len(t, values3, 1)
	assert.Equal(t, []byte(nil), values3[0])
}

func TestKvdbMemSet(t *testing.T) {
	os.RemoveAll(store_cfg2.DbPath)
	store := New(store_cfg2).(*Store)
	assert.NotNil(t, store)

	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{[]byte("mk1"), []byte("v1")})
	kv = append(kv, &types.KeyValue{[]byte("mk2"), []byte("v2")})
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(t, err)
	keys := [][]byte{[]byte("mk1"), []byte("mk2")}
	get1 := &types.StoreGet{hash, keys}

	values := store.Get(get1)
	assert.Len(t, values, 2)

	actHash, _ := store.Commit(&types.ReqHash{hash})
	assert.Equal(t, hash, actHash)

	notExistHash, _ := store.Commit(&types.ReqHash{drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
}

func TestKvdbRollback(t *testing.T) {
	os.RemoveAll(store_cfg3.DbPath)
	store := New(store_cfg3).(*Store)
	assert.NotNil(t, store)

	var kv []*types.KeyValue
	kv = append(kv, &types.KeyValue{[]byte("mk1"), []byte("v1")})
	kv = append(kv, &types.KeyValue{[]byte("mk2"), []byte("v2")})
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(t, err)
	keys := [][]byte{[]byte("mk1"), []byte("mk2")}
	get1 := &types.StoreGet{hash, keys}
	values := store.Get(get1)
	assert.Len(t, values, 2)

	actHash, _ := store.Rollback(&types.ReqHash{hash})
	assert.Equal(t, hash, actHash)

	notExistHash, _ := store.Rollback(&types.ReqHash{drivers.EmptyRoot[:]})
	assert.Nil(t, notExistHash)
}

var checkKVResult []*types.KeyValue

func checkKV(k, v []byte) bool {
	checkKVResult = append(checkKVResult,
		&types.KeyValue{k, v})
	//mlog.Debug("checkKV", "key", string(k), "value", string(v))
	return false
}

//生成随机字符串
func GetRandomString(lenth int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < lenth; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func BenchmarkGet(b *testing.B) {
	os.RemoveAll(store_cfg6.DbPath)
	store := New(store_cfg6).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < b.N; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{[]byte(string(key)), []byte(string(value))})
	}
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	hash, err := store.Set(datas, true)
	assert.Nil(b, err)
	getData := &types.StoreGet{
		hash,
		keys}

	start := time.Now()
	b.ResetTimer()
	values := store.Get(getData)
	end := time.Now()
	fmt.Println("mpt BenchmarkGet cost time is", end.Sub(start), "num is", b.N)
	assert.Len(b, values, b.N)
	b.StopTimer()
}

func BenchmarkSet(b *testing.B) {
	os.RemoveAll(store_cfg7.DbPath)
	store := New(store_cfg7).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < b.N; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{[]byte(string(key)), []byte(string(value))})
	}
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	start := time.Now()
	b.ResetTimer()
	hash, err := store.Set(datas, true)
	assert.Nil(b, err)
	assert.NotNil(b, hash)
	end := time.Now()
	fmt.Println("mpt BenchmarkSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkMemSet(b *testing.B) {
	os.RemoveAll(store_cfg8.DbPath)
	store := New(store_cfg8).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < b.N; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{[]byte(string(key)), []byte(string(value))})
	}
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	start := time.Now()
	b.ResetTimer()
	hash, err := store.MemSet(datas, true)
	assert.Nil(b, err)
	assert.NotNil(b, hash)
	end := time.Now()
	fmt.Println("mpt BenchmarkMemSet cost time is", end.Sub(start), "num is", b.N)
}

func BenchmarkCommit(b *testing.B) {
	os.RemoveAll(store_cfg9.DbPath)
	store := New(store_cfg9).(*Store)
	assert.NotNil(b, store)

	var kv []*types.KeyValue
	var key string
	var value string
	var keys [][]byte

	for i := 0; i < b.N; i++ {
		key = GetRandomString(MaxKeylenth)
		value = fmt.Sprintf("v%d", i)
		keys = append(keys, []byte(string(key)))
		kv = append(kv, &types.KeyValue{[]byte(string(key)), []byte(string(value))})
	}
	datas := &types.StoreSet{
		drivers.EmptyRoot[:],
		kv,
		0}
	hash, err := store.MemSet(datas, true)
	assert.Nil(b, err)
	req := &types.ReqHash{
		Hash: hash,
	}

	start := time.Now()
	b.ResetTimer()
	_, err = store.Commit(req)
	assert.NoError(b, err, "NoError")
	end := time.Now()
	fmt.Println("mpt BenchmarkCommit cost time is", end.Sub(start), "num is", b.N)
	b.StopTimer()
}