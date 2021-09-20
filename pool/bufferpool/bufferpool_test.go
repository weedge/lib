package bufferpool

import (
	"bytes"
	"testing"
)

var count = 5
var str = `{"errNo":0,"errStr":"fsd","data":{"105229_0_10":"{\"examInfo\":{\"bindInfo\":{\"bindId\":105229,\"bindType\":0,\"createTime\":1535797800,\"examId\":30188,\"examType\":10,\"operatorName\":\"鲍伟\",\"operatorUid\":18,\"props\":{\"duration\":0,\"maxTryNum\":0,\"passScore\":0},\"relationId\":28418,\"updateTime\":1537023831,\"userKv\":{\"examStatus\":2,\"startTime\":1537023586}},\"examInfo\":\"{\\\"examId\\\":30188,\\\"examType\\\":10,\\\"title\\\":\\\"\\\\u79cb1\\\\u9ad8\\\\u4e09\\\\u7269\\\\u7406\\\\u540c\\\\u6b65(\\\\u5c16\\\\u7aef\\\\u57f9\\\\u4f18\\\\u3001\\\\u5f3a\\\\u5316\\\\u63d0\\\\u5347)\\\\u7b2c2\\\\u8bb2\\\\u5802\\\\u5802\\\\u6d4b\\\",\\\"tidList\\\":{\\\"376085497\\\":{\\\"score\\\":50,\\\"type\\\":2},\\\"375574847\\\":{\\\"score\\\":50,\\\"type\\\":1}},\\\"totalScore\\\":100,\\\"props\\\":[],\\\"userKv\\\":[],\\\"grade\\\":7,\\\"subject\\\":4,\\\"ruleInfo\\\":{\\\"duration\\\":0,\\\"passScore\\\":0,\\\"maxTryNum\\\":0},\\\"extData\\\":[]}\",\"path\":\"c:0-l:0-cpu:0\"},\"questionList\":{}}","105229_0_13":"{\"examInfo\":{},\"questionList\":{}}","105229_0_7":"{\"examInfo\":{\"bindInfo\":{\"bindId\":105229,\"bindType\":0,\"createTime\":1535797998,\"examId\":100724,\"examType\":7,\"operatorName\":\"鲍伟\",\"operatorUid\":18,\"props\":{\"duration\":0,\"maxTryNum\":0,\"passScore\":0},\"relationId\":38712,\"updateTime\":1535797998,\"userKv\":[]},\"examInfo\":\"{\\\"examId\\\":100724,\\\"examType\\\":7,\\\"title\\\":\\\"\\\\u79cb1\\\\u9ad8\\\\u4e09\\\\u7269\\\\u7406\\\\u540c\\\\u6b65(\\\\u5c16\\\\u7aef\\\\u57f9\\\\u4f18\\\\u3001\\\\u5f3a\\\\u5316\\\\u63d0\\\\u5347)\\\\u7b2c2\\\\u8bb2\\\\u8bfe\\\\u540e\\\\u4f5c\\\\u4e1a\\\",\\\"tidList\\\":{\\\"375479289\\\":{\\\"score\\\":0,\\\"type\\\":2},\\\"375574847\\\":{\\\"score\\\":0,\\\"type\\\":1}},\\\"totalScore\\\":0,\\\"props\\\":[],\\\"userKv\\\":[],\\\"grade\\\":7,\\\"subject\\\":4,\\\"ruleInfo\\\":{\\\"duration\\\":0,\\\"passScore\\\":0,\\\"maxTryNum\\\":0},\\\"extData\\\":[]}\",\"path\":\"c:0-l:0-cpu:0\"},\"questionList\":{}}"}}`

func Benchmark_TestStringAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < count; j++ {
			s = s + str
		}
	}
}

func Benchmark_TestBufferPool(b *testing.B) {
	b.StopTimer() //调用该函数停止压力测试的时间计数
	var buffers = NewBufferPool(516)
	b.StartTimer() //重新开始时间
	for i := 0; i < b.N; i++ {
		buf := buffers.Get()
		for j := 0; j < count; j++ {
			buf.WriteString(str)
		}
		buffers.Put(buf)
	}
}

func Benchmark_TestBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBufferString("")
		for j := 0; j < count; j++ {
			buf.WriteString(str)
		}
	}
}

func BenchmarkJI_TestBufferPool_Parallel(b *testing.B) {
	b.StopTimer() //调用该函数停止压力测试的时间计数
	var buffers = NewBufferPool(516)
	b.StartTimer() //重新开始时间
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := buffers.Get()
			for j := 0; j < count; j++ {
				buf.WriteString(str)
			}
			buffers.Put(buf)
		}
	})
}

func BenchmarkJI_TestBuffer_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bytes.NewBufferString("")
			for j := 0; j < count; j++ {
				buf.WriteString(str)
			}
		}
	})
}

func BenchmarkJI_TestStringAppend_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := ""
			for j := 0; j < count; j++ {
				s = s + str
			}
		}
	})
}
