package test

import (
	"testing"

	"github.com/zengzhengrong/request/curl"
	"github.com/zengzhengrong/request/opts/client"
)

func TestTenGETWithoutReused(t *testing.T) {
	for i := 0; i < 10; i++ {
		resp := curl.GET("http://127.0.0.1:8081/get", query, header)
		if resp.Err != nil {
			panic(resp.Err)
		}
	}
}

func TestTenGETWithReused(t *testing.T) {
	client := client.NewClient(
		client.WithDefault(),
	)
	for i := 0; i < 10; i++ {
		resp := client.GET("http://127.0.0.1:8081/get", query, header)
		if resp.Err != nil {
			panic(resp.Err)
		}
	}
}

// BenchmarkGETWithoutReused 性能测试 没有复用会话
func BenchmarkGETWithoutReused(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := curl.GET("http://127.0.0.1:8081/get", query, header)
		if resp.Err != nil {
			panic(resp.Err)
		}
	}

}

// BenchmarkGETReused 性能测试 复用会话
func BenchmarkGETReused(b *testing.B) {
	client := client.NewClient(
		client.WithDefault(),
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := client.GET("http://127.0.0.1:8081/get", query, header)
		if resp.Err != nil {
			panic(resp.Err)
		}
	}

}
