package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	u "github.com/ipfs/go-ipfs-util"
	format "github.com/ipfs/go-ipld-format"
	multihash "github.com/multiformats/go-multihash"
)

func main() {
	helloHash := "QmZULkCELmmk5XNfCgTnCyFgAVxBRBXyDHGGMVoLFLiXEN"
	helloMH, _ := multihash.FromB58String(helloHash)
	helloSUM := u.Hash([]byte("hello"))
	helloSUM2 := u.Hash([]byte{10, 12, 8, 2, 18, 6, 104, 101, 108, 108, 111, 10, 24, 6})
	fmt.Println(helloSUM2.B58String())

	fmt.Println([]byte{10, 12, 8, 2, 18, 6, 104, 101, 108, 108, 111, 10, 24, 6})
	blockH := blocks.NewBlock([]byte{10, 12, 8, 2, 18, 6, 104, 101, 108, 108, 111, 10, 24, 6})
	fmt.Println(blockH)
	fmt.Println(blockH.Multihash())
	fmt.Println([]byte("hello"))

	bbb := blocks.NewBlock([]byte("hello"))
	fmt.Println(bbb.RawData())

	// 10 12 8 2 18 6         // 10 24 6

	fmt.Println("---------")
	fmt.Println("---------")
	fmt.Println("---------")
	fmt.Println("---------")
	fmt.Println("---------")
	fmt.Println("---------")
	fmt.Println(helloMH)
	helloC := cid.NewCidV0(helloMH)
	fmt.Println(helloC.Hash())
	fmt.Println("---------")
	fmt.Println(helloSUM)

	fmt.Println("--------------------------")
	helloS := cid.NewCidV0(helloSUM)
	fmt.Println(helloS.Hash())
	fmt.Println("----")

	bb := blocks.NewBlock([]byte("hello"))
	fmt.Println("block EXP")
	fmt.Println(bb.Multihash())
	fmt.Println(bb.Cid())
	fmt.Println(bb.RawData())
	fmt.Println(bb.Loggable())

	rawHash := "QmevaAiHWVwEqwm3wqQyhqVbwJQ5QT1q47su9xKTgVMrtF"
	mh, err := multihash.FromB58String(rawHash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("mh")
	fmt.Println(mh)
	decoded, err := multihash.Decode(mh)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(decoded)

	resp, err := http.Get("https://ipfs.infura.io/ipfs/QmevaAiHWVwEqwm3wqQyhqVbwJQ5QT1q47su9xKTgVMrtF")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	resp2, err := http.Get("http://localhost:5001/api/v0/dag/get?arg=QmevaAiHWVwEqwm3wqQyhqVbwJQ5QT1q47su9xKTgVMrtF")
	if err != nil {
		log.Fatal(err)
	}
	defer resp2.Body.Close()
	body2, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		log.Fatal(err)
	}

	newHash := base58Sha2256Multihash(body)

	data, _ := base64.StdEncoding.DecodeString(payload)
	hash2 := base58Sha2256Multihash(data[5 : len(data)-3])
	fmt.Println(len(body))
	fmt.Println(len(data[5:]))
	fmt.Printf("source hash: %v\n", rawHash)
	fmt.Printf("new hash   : %v\n", newHash)
	fmt.Printf("new hash2  : %v\n", hash2)

	basicBlock := blocks.NewBlock(body)
	mh2 := basicBlock.Multihash()
	fmt.Println(mh2.B58String())
	basicBlock2 := blocks.NewBlock(body2)
	mh3 := basicBlock2.Multihash()
	fmt.Println(mh2)
	fmt.Println(mh2.B58String())
	fmt.Println(mh3)
	fmt.Println(mh3.B58String())

	data6 := body
	hash6, err := multihash.Sum(data6, multihash.SHA2_256, -1)
	if err != nil {
		log.Fatal(err)
	}

	c1 := cid.NewCidV0(hash6)
	fmt.Println("c1")
	fmt.Println(c1.Hash())

	fmt.Println("basic block")
	fmt.Println(basicBlock)

	node, err := format.Decode(basicBlock2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)

	// Create a cid manually by specifying the 'prefix' parameters
	pref := cid.Prefix{
		Version:  0,
		Codec:    cid.Raw,
		MhType:   uint64(18),
		MhLength: -1, // default length
	}

	// And then feed it some data
	c, err := pref.Sum(body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created CID: ", c)

}

func base58Sha2256Multihash(b []byte) string {
	code := uint64(18)
	length := -1
	mh, _ := multihash.Sum(b, code, length)
	return mh.B58String()
}

var payload = "CAISpkk8IWRvY3R5cGUgaHRtbD4KPGh0bWw+CjxoZWFkPgo8dGl0bGU+SGFyZGJpbjwvdGl0bGU+CjxsaW5rIHJlbD0iaWNvbiIgaHJlZj0iaW1nL2gucG5nIj4KPGxpbmsgcmVsPSJzdHlsZXNoZWV0IiBocmVmPSJjc3MvYm9vdHN0cmFwLm1pbi5jc3MiPgo8bGluayByZWw9InN0eWxlc2hlZXQiIGhyZWY9ImNzcy9oYXJkYmluLmNzcyI+CjwvaGVhZD4KPGJvZHk+CjxkaXYgaWQ9ImNvbnRhaW5lciI+CjxkaXYgaWQ9Im5hdiI+CiAgICA8ZGl2IGlkPSJ0aXRsZS1kaXYiPgogICAgICAgIDxoMT5oYXJkYmluLjwvaDE+CiAgICA8L2Rpdj4KCiAgICA8ZGl2IGlkPSJzdGF0dXMtZGl2Ij4KICAgICAgICA8c3BhbiBpZD0ic3RhdHVzIj5Mb2FkaW5nLi4uPC9zcGFuPgogICAgPC9kaXY+CgogICAgPGRpdiBpZD0iY29udHJvbHMtZGl2Ij4KICAgICAgICA8YnV0dG9uIGNsYXNzPSJidG4gYnRuLWRlZmF1bHQiIGlkPSJhYm91dCI+QWJvdXQ8L2J1dHRvbj4gfAogICAgICAgIDxhIGNsYXNzPSJidG4gYnRuLWRlZmF1bHQiIGhyZWY9Ii4iPk5ldzwvYT4KICAgICAgICA8YnV0dG9uIGNsYXNzPSJidG4gYnRuLXByaW1hcnkiIGlkPSJ0b3Atc2F2ZSI+UHVibGlzaDwvYnV0dG9uPgogICAgICAgIDxidXR0b24gY2xhc3M9ImJ0biBidG4tcHJpbWFyeSIgaWQ9InRvcC1lZGl0IiBzdHlsZT0iZGlzcGxheTpub25lIj5FZGl0PC9idXR0b24+CiAgICA8L2Rpdj4KCjwvZGl2PgoKPGRpdiBzdHlsZT0iY2xlYXI6Ym90aCI+PC9kaXY+Cgo8ZGl2IGlkPSJpbnB1dC1kaXYiPgogICAgPGRpdiBpZD0idGV4dGFyZWEtZGl2Ij48dGV4dGFyZWEgYXV0b2ZvY3VzIHJvd3M9IjEwIiBjbGFzcz0iZm9ybS1jb250cm9sIiBpZD0iaW5wdXQiIHBsYWNlaG9sZGVyPSJJbnB1dCB5b3VyIHRleHQgaGVyZSwgdGhlbiBjbGljayBQdWJsaXNoLiI+PC90ZXh0YXJlYT48L2Rpdj4KPC9kaXY+CjwvZGl2PgoKPGRpdiBpZD0iYWJvdXQtbW9kYWwiIGNsYXNzPSJtb2RhbCBmYWRlIiB0YWJpbmRleD0iLTEiIHJvbGU9ImRpYWxvZyI+CiAgICA8ZGl2IGNsYXNzPSJtb2RhbC1kaWFsb2cgd2lkZS1tb2RhbCIgcm9sZT0iZG9jdW1lbnQiPgogICAgICAgIDxkaXYgY2xhc3M9Im1vZGFsLWNvbnRlbnQiPgogICAgICAgICAgPGRpdiBjbGFzcz0ibW9kYWwtaGVhZGVyIj4KICAgICAgICAgICAgPGJ1dHRvbiB0eXBlPSJidXR0b24iIGNsYXNzPSJjbG9zZSIgZGF0YS1kaXNtaXNzPSJtb2RhbCIgYXJpYS1sYWJlbD0iQ2xvc2UiPjxzcGFuIGFyaWEtaGlkZGVuPSJ0cnVlIj4mdGltZXM7PC9zcGFuPjwvYnV0dG9uPgogICAgICAgICAgICA8aDQgY2xhc3M9Im1vZGFsLXRpdGxlIj5BYm91dDwvaDQ+CiAgICAgICAgICA8L2Rpdj4KICAgICAgICAgIDxkaXYgY2xhc3M9Im1vZGFsLWJvZHkiIGlkPSJhYm91dC1ib2R5Ij4KICAgICAgICAgICAgPHA+SGFyZGJpbiBpcyBhbiBlbmNyeXB0ZWQgcGFzdGViaW4gdXNpbmcgSVBGUy4gSXQgd2FzIGNyZWF0ZWQgYnkgPGEgaHJlZj0iaHR0cDovL2luY29oZXJlbmN5LmNvLnVrLyI+SmFtZXMgU3RhbmxleTwvYT4uPC9wPgogICAgICAgICAgICA8cD5Vbmxpa2Ugb3RoZXIgcGFzdGViaW5zLCBIYXJkYmluIGRvZXMgbm90IHJlcXVpcmUgeW91IHRvIHRydXN0IGFueSBzZXJ2ZXIuIFlvdSBjYW4gcnVuIGEgbG9jYWwgPGEgaHJlZj0iaHR0cHM6Ly9pcGZzLmlvLyI+SVBGUzwvYT4KICAgICAgICAgICAgZ2F0ZXdheSBhbmQgdGhlbiB5b3UgY2FuIGFsd2F5cyBiZSBjZXJ0YWluIHRoYXQgbm8gcmVtb3RlIHNlcnZlciBpcyBhYmxlIHRvIG1vZGlmeSB0aGUgY29kZSB5b3UncmUgcnVubmluZy4gSW4gcGFydGljdWxhciwgdGhpcwogICAgICAgICAgICBtZWFucyBubyByZW1vdGUgc2VydmVyIGlzIGFibGUgdG8gaW5zZXJ0IG1hbGljaW91cyBjb2RlIHRvIGV4ZmlsdHJhdGUgdGhlIGNvbnRlbnQgb2YgeW91ciBwYXN0ZXMuPC9wPgogICAgICAgICAgICA8cD5UaGVyZSBpcyBhIHB1YmxpYyB3cml0YWJsZSBnYXRld2F5IGF2YWlsYWJsZSBhdCA8YSBocmVmPSJodHRwczovL2hhcmRiaW4uY29tLyI+aGFyZGJpbi5jb208L2E+IHdoaWNoIGFsbG93cyBjcmVhdGlvbiBvZiBwYXN0ZXMKICAgICAgICAgICAgd2l0aG91dCBydW5uaW5nIGEgbG9jYWwgZ2F0ZXdheS48L3A+CiAgICAgICAgICAgIDxwPklmIHlvdSB3YW50IHRvIGxlYXJuIG1vcmUsIHBsZWFzZSBzZWUgPGEgaHJlZj0iUkVBRE1FLm1kIj5SRUFETUUubWQ8L2E+LjwvcD4KICAgICAgICAgICAgPHA+SWYgeW91IHdhbnQgdG8gY29udHJpYnV0ZSB0byBIYXJkYmluIGRldmVsb3BtZW50LCBwbGVhc2Ugc2VlIHRoZSA8YSBocmVmPSJodHRwczovL2dpdGh1Yi5jb20vamVzL2hhcmRiaW4vIj5naXRodWIgcmVwbzwvYT4uPC9wPgogICAgICAgICAgPC9kaXY+CiAgICAgICAgICA8ZGl2IGNsYXNzPSJtb2RhbC1mb290ZXIiPgogICAgICAgICAgICA8YnV0dG9uIHR5cGU9ImJ1dHRvbiIgY2xhc3M9ImJ0biBidG4tZGVmYXVsdCIgZGF0YS1kaXNtaXNzPSJtb2RhbCI+Q2xvc2U8L2J1dHRvbj4KICAgICAgICAgIDwvZGl2PgogICAgICAgIDwvZGl2PgogICAgPC9kaXY+CjwvZGl2PgoKPGRpdiBpZD0ibW9kYWwiIGNsYXNzPSJtb2RhbCBmYWRlIiB0YWJpbmRleD0iLTEiIHJvbGU9ImRpYWxvZyI+CiAgICA8ZGl2IGNsYXNzPSJtb2RhbC1kaWFsb2cgd2lkZS1tb2RhbCIgcm9sZT0iZG9jdW1lbnQiPgogICAgICAgIDxkaXYgY2xhc3M9Im1vZGFsLWNvbnRlbnQiPgogICAgICAgICAgPGRpdiBjbGFzcz0ibW9kYWwtaGVhZGVyIj4KICAgICAgICAgICAgPGJ1dHRvbiB0eXBlPSJidXR0b24iIGNsYXNzPSJjbG9zZSIgZGF0YS1kaXNtaXNzPSJtb2RhbCIgYXJpYS1sYWJlbD0iQ2xvc2UiPjxzcGFuIGFyaWEtaGlkZGVuPSJ0cnVlIj4mdGltZXM7PC9zcGFuPjwvYnV0dG9uPgogICAgICAgICAgICA8aDQgY2xhc3M9Im1vZGFsLXRpdGxlIiBpZD0ibW9kYWwtdGl0bGUiPk1vZGFsIHRpdGxlPC9oND4KICAgICAgICAgIDwvZGl2PgogICAgICAgICAgPGRpdiBjbGFzcz0ibW9kYWwtYm9keSIgaWQ9Im1vZGFsLWJvZHkiPgogICAgICAgICAgICBNb2RhbCBib2R5LgogICAgICAgICAgPC9kaXY+CiAgICAgICAgICA8ZGl2IGNsYXNzPSJtb2RhbC1mb290ZXIiPgogICAgICAgICAgICA8YnV0dG9uIHR5cGU9ImJ1dHRvbiIgY2xhc3M9ImJ0biBidG4tZGVmYXVsdCIgZGF0YS1kaXNtaXNzPSJtb2RhbCI+Q2xvc2U8L2J1dHRvbj4KICAgICAgICAgIDwvZGl2PgogICAgICAgIDwvZGl2PgogICAgPC9kaXY+CjwvZGl2PgoKPHNjcmlwdCBzcmM9ImpzL2pxdWVyeS0zLjIuMS5taW4uanMiPjwvc2NyaXB0Pgo8c2NyaXB0IHNyYz0ianMvYm9vdHN0cmFwLm1pbi5qcyI+PC9zY3JpcHQ+CjxzY3JpcHQgc3JjPSJqcy9hZXMuanMiPjwvc2NyaXB0Pgo8c2NyaXB0IHNyYz0ianMvanNibi5qcyI+PC9zY3JpcHQ+CjxzY3JpcHQgc3JjPSJqcy9qc2JuMi5qcyI+PC9zY3JpcHQ+CjxzY3JpcHQgc3JjPSJqcy9iYXNlNTguanMiPjwvc2NyaXB0Pgo8c2NyaXB0IHNyYz0ianMvaGFyZGJpbi5qcyI+PC9zY3JpcHQ+CjxzY3JpcHQgc3JjPSJqcy9zaG93ZG93bi5taW4uanMiPjwvc2NyaXB0Pgo8c2NyaXB0IHR5cGU9InRleHQvamF2YXNjcmlwdCI+CnZhciBjaGVja2VkX3dyaXRhYmlsaXR5ID0gZmFsc2U7Cgp2YXIgZGVjcnlwdGlvbl9rZXkgPSAnJzsKCmZ1bmN0aW9uIHNldF9zdGF0dXModHh0KSB7CiAgICAkKCcjc3RhdHVzJykudGV4dCh0eHQpOwp9CgpmdW5jdGlvbiBtb2RhbCh0aXRsZSwgYm9keWh0bWwpIHsKICAgICQoJyNtb2RhbC10aXRsZScpLnRleHQodGl0bGUpOwogICAgJCgnI21vZGFsLWJvZHknKS5odG1sKGJvZHlodG1sKTsKICAgICQoJyNtb2RhbCcpLm1vZGFsKCdzaG93Jyk7Cn0KCmZ1bmN0aW9uIGNoZWNrX3dyaXRhYmlsaXR5KCkgewogICAgY2hlY2tlZF93cml0YWJpbGl0eSA9IHRydWU7CiAgICBzZXRfc3RhdHVzKCJDaGVja2luZyBnYXRld2F5IGZvciB3cml0YWJpbGl0eS4uLiIpOwogICAgd3JpdGUoInRlc3RpbmcgMTIzIiwgZnVuY3Rpb24oaGFzaCkgewogICAgICAgIGlmIChoYXNoKSB7CiAgICAgICAgICAgIHNldF9zdGF0dXMoIiIpOwogICAgICAgIH0gZWxzZSBpZiAoaXNfbG9jYWxfZ2F0ZXdheSgpKSB7CiAgICAgICAgICAgIG1vZGFsKCJJUEZTIEdhdGV3YXkgUHJvYmxlbSIsICI8cD5JdCBsb29rcyBsaWtlIHlvdSdyZSBhY2Nlc3NpbmcgSGFyZGJpbiBvdmVyIGEgbG9jYWwgZ2F0ZXdheS4gVGhhdCdzIGdvb2QhIFRoYXQncyB0aGUgc2FmZXN0IHdheS4gQnV0IHlvdXIgZ2F0ZXdheSBpcyBub3QgY3VycmVudGx5IHdyaXRhYmxlLCB3aGljaCBtZWFucyB5b3Ugd29uJ3QgYmUgYWJsZSB0byBzYXZlIHlvdXIgd29yay48L3A+PHA+S2lsbCBpdCBhbmQgcmVsYXVuY2ggd2l0aCA8dHQ+LS13cml0YWJsZTwvdHQ+IGlmIHlvdSB3YW50IHRvIHNhdmUgeW91ciB3b3JrOjwvcD48cD48cHJlPjxjb2RlPiQgaXBmcyBkYWVtb24gLS13cml0YWJsZTwvY29kZT48L3ByZT48L3A+Iik7CiAgICAgICAgICAgIHNldF9zdGF0dXMoIkVycm9yOiBJUEZTIGdhdGV3YXkgaXMgbm90IHdyaXRhYmxlLiIpOwogICAgICAgIH0gZWxzZSB7CiAgICAgICAgICAgIHZhciBwYXRod2l0aGZyYWcgPSB3aW5kb3cubG9jYXRpb24ucGF0aG5hbWUgKyB3aW5kb3cubG9jYXRpb24uaGFzaDsKICAgICAgICAgICAgLy8gVE9ETzogY2hlY2sgaWYgd2UgY2FuIGZldGNoIGlwZnMgY29udGVudCBmcm9tIGh0dHA6Ly9sb2NhbGhvc3Q6ODA4MC8gYW5kIG9mZmVyIGl0IGFzIGFuIG9wdGlvbiBpZiBzbwogICAgICAgICAgICBtb2RhbCgiSVBGUyBHYXRld2F5IFByb2JsZW0iLCAiPHA+VGhpcyBJUEZTIGdhdGV3YXkgaXMgbm90IHdyaXRhYmxlLCB3aGljaCBtZWFucyB5b3Ugd29uJ3QgYmUgYWJsZSB0byBzYXZlIHlvdXIgd29yay48L3A+PHA+SWYgeW91IHdhbnQgdG8gc2F2ZSB5b3VyIHdvcmssIHlvdSBjYW4gZWl0aGVyOjwvcD48cD4xLiB2aWV3IHRoaXMgb24gdGhlIHB1YmxpYyB3cml0YWJsZSBnYXRld2F5IGF0IDxhIGhyZWY9XCJodHRwczovL2hhcmRiaW4uY29tIiArIHBhdGh3aXRoZnJhZyArICJcIj5oYXJkYmluLmNvbTwvYT4sIG9yPC9wPjxwPjIuIDxhIGhyZWY9XCJodHRwczovL2lwZnMuaW8vZG9jcy9pbnN0YWxsL1wiPmluc3RhbGwgSVBGUzwvYT4sIHJ1biBhIGxvY2FsIG5vZGUgd2l0aCA8dHQ+aXBmcyBkYWVtb24gLS13cml0YWJsZTwvdHQ+LCBhbmQgdGhlbiB2aWV3IHRoaXMgb24geW91ciBsb2NhbCBub2RlIGF0IDxhIGhyZWY9XCJodHRwOi8vbG9jYWxob3N0OjgwODAiICsgcGF0aHdpdGhmcmFnICsgIlwiPmxvY2FsaG9zdDo4MDgwPC9hPi48L3A+Iik7CiAgICAgICAgICAgIHNldF9zdGF0dXMoIkVycm9yOiBJUEZTIGdhdGV3YXkgaXMgbm90IHdyaXRhYmxlLiIpOwogICAgICAgIH0KICAgIH0pOwp9CgpmdW5jdGlvbiByZW5kZXIoY29udGVudCkgewogICAgJCgnI2lucHV0JykudmFsKGNvbnRlbnQpOwogICAgJCgnI2lucHV0JykucHJvcCgncmVhZG9ubHknLHRydWUpOwoKICAgICQoJyN0b3Atc2F2ZScpLmhpZGUoKTsKICAgICQoJyN0b3AtZWRpdCcpLnNob3coKTsKfQoKZnVuY3Rpb24gdW5yZW5kZXIoKSB7CiAgICAkKCcjaW5wdXQnKS5wcm9wKCdyZWFkb25seScsZmFsc2UpOwoKICAgICQoJyN0b3Atc2F2ZScpLnNob3coKTsKICAgICQoJyN0b3AtZWRpdCcpLmhpZGUoKTsKfQoKZnVuY3Rpb24gc2hvd19waW5faW5zdHJ1Y3Rpb25zKCkgewogICAgdmFyIHBhdGh3aXRoZnJhZyA9IHdpbmRvdy5sb2NhdGlvbi5wYXRobmFtZSArICcjJyArIGRlY3J5cHRpb25fa2V5OwogICAgdmFyIGZ1bGxsb2NhdGlvbiA9IHdpbmRvdy5sb2NhdGlvbi5ocmVmOwogICAgdmFyIGhhc2ggPSB3aW5kb3cubG9jYXRpb24ucGF0aG5hbWU7CiAgICBoYXNoID0gaGFzaC5yZXBsYWNlKCcvaXBmcy8nLCAnJyk7CiAgICBoYXNoID0gaGFzaC5yZXBsYWNlKCcvJywgJycpOwogICAgZnVsbGxvY2F0aW9uID0gZnVsbGxvY2F0aW9uLnJlcGxhY2UoJy1maXJzdHZpZXcnLCAnJyk7CiAgICB2YXIgaXNfaGFyZGJpbmNvbSA9IHdpbmRvdy5sb2NhdGlvbi5ob3N0bmFtZSA9PSAnaGFyZGJpbi5jb20nOwogICAgbW9kYWwoIkNvbnRlbnQgcHVibGlzaGVkIiwgIjxwPkNvbmdyYXR1bGF0aW9ucyEgWW91ciBjb250ZW50IGhhcyBiZWVuIHB1Ymxpc2hlZCB0byB0aGUgSVBGUyBnYXRld2F5LiBJdCBpcyBub3cgcmVhY2hhYmxlIGJ5IGFueSBub2RlIG9uIHRoZSBJUEZTIG5ldHdvcmsuIFNoYXJlIHRoZSBmb2xsb3dpbmcgVVJMIHRvIHNoYXJlIHRoZSBjb250ZW50OjwvcD48cD48YSBzdHlsZT1cIndvcmQtd3JhcDpicmVhay13b3JkXCIgaHJlZj0iICsgZnVsbGxvY2F0aW9uICsgIj4iICsgZnVsbGxvY2F0aW9uICsgIjwvYT48L3A+IiArICghaXNfaGFyZGJpbmNvbSA/ICI8cD5PciBvbiB0aGUgaGFyZGJpbi5jb20gcHVibGljIGdhdGV3YXk6IDxhIHN0eWxlPVwid29yZC13cmFwOmJyZWFrLXdvcmRcIiBocmVmPVwiaHR0cHM6Ly9oYXJkYmluLmNvbSIgKyBwYXRod2l0aGZyYWcgKyAiXCI+aHR0cHM6Ly9oYXJkYmluLmNvbSIgKyBwYXRod2l0aGZyYWcgKyAiPC9hPjwvcD4iIDogJycpICsgIjxwPlRoZSBJUEZTIGhhc2ggaXMgPGI+IiArIGhhc2ggKyAiPC9iPiBhbmQgdGhlIGRlY3J5cHRpb24ga2V5IGlzIDxiPiIgKyBkZWNyeXB0aW9uX2tleSArICI8L2I+LjwvcD48cD5Db250ZW50IG9uIElQRlMgaXMgbm90IHBlcnNpc3RlbnQgYW5kIHdpbGwgZXZlbnR1YWxseSBkaXNhcHBlYXIgZnJvbSB0aGUgSVBGUyBuZXR3b3JrIGlmIGl0IGlzIG5vdCBwaW5uZWQgYW55d2hlcmUgKGVxdWl2YWxlbnQgdG8gXCJzZWVkaW5nXCIgaW4gYml0dG9ycmVudCkuIFRvIG1ha2UgdGhlIGNvbnRlbnQgcGVyc2lzdGVudCwgeW91IGNhbiBlaXRoZXIgcGluIGl0IG9uIGFuIElQRlMgbm9kZSB5b3UgY29udHJvbDo8L3A+PHA+PHR0PiQgaXBmcyBwaW4gYWRkICIgKyBoYXNoICsgIjwvdHQ+PC9wPjxwPk9yIHVzZSBhIHNlcnZpY2UgbGlrZSBJUEZTc3RvcmUgdG8gcGluIGl0IGZvciB5b3U6PC9wPjxhIGhyZWY9XCJodHRwczovL2lwZnNzdG9yZS5pdC9zdWJtaXQucGhwP2hhc2g9IiArIGhhc2ggKyAiXCIgY2xhc3M9XCJidG4gYnRuLXByaW1hcnlcIj5QaW4gb24gSVBGU3N0b3JlPC9hPiIpOwp9CgpmdW5jdGlvbiBsb2FkX2NvbnRlbnQoKSB7CiAgICBpZiAod2luZG93LmxvY2F0aW9uLmhhc2ggJiYgd2luZG93LmxvY2F0aW9uLmhhc2ggIT0gJyNhYm91dCcpIHsKICAgICAgICBzZXRfc3RhdHVzKCJMb2FkaW5nIGVuY3J5cHRlZCBjb250ZW50Li4uIik7CiAgICAgICAga2V5ID0gd2luZG93LmxvY2F0aW9uLmhhc2g7CiAgICAgICAga2V5ID0ga2V5LnJlcGxhY2UoJy1maXJzdHZpZXcnLCcnKTsKICAgICAgICBrZXkgPSBrZXkucmVwbGFjZSgnIycsJycpOwogICAgICAgIGRlY3J5cHRpb25fa2V5ID0ga2V5OwogICAgICAgICQuYWpheCh7CiAgICAgICAgICAgIHVybDogImNvbnRlbnQiLAogICAgICAgICAgICBzdWNjZXNzOiBmdW5jdGlvbihkYXRhKSB7CiAgICAgICAgICAgICAgICAvLyBUT0RPOiBzaG93IGFuIGVycm9yIGlmIHdlIGNvdWxkbid0IGRlY3J5cHQgdGhlIGNvbnRlbnQ/CiAgICAgICAgICAgICAgICB2YXIgcGxhaW4gPSBkZWNyeXB0KGRhdGEsa2V5KTsKICAgICAgICAgICAgICAgIGlmIChwbGFpbikgewogICAgICAgICAgICAgICAgICAgIHJlbmRlcihwbGFpbik7CiAgICAgICAgICAgICAgICAgICAgaWYgKHdpbmRvdy5sb2NhdGlvbi5oYXNoLmluZGV4T2YoJ2ZpcnN0dmlldycpICE9IC0xKQogICAgICAgICAgICAgICAgICAgICAgICBzaG93X3Bpbl9pbnN0cnVjdGlvbnMoKTsKICAgICAgICAgICAgICAgICAgICBoaXN0b3J5LnJlcGxhY2VTdGF0ZSh1bmRlZmluZWQsIHVuZGVmaW5lZCwgJyMnICsgZGVjcnlwdGlvbl9rZXkpOwogICAgICAgICAgICAgICAgfQogICAgICAgICAgICAgICAgc2V0X3N0YXR1cygiIik7CiAgICAgICAgICAgIH0sCiAgICAgICAgICAgIGVycm9yOiBmdW5jdGlvbigpIHsKICAgICAgICAgICAgICAgIGNoZWNrX3dyaXRhYmlsaXR5KCk7CiAgICAgICAgICAgIH0sCiAgICAgICAgICAgIHRpbWVvdXQ6IGZ1bmN0aW9uKCkgewogICAgICAgICAgICAgICAgY2hlY2tfd3JpdGFiaWxpdHkoKTsKICAgICAgICAgICAgfQogICAgICAgIH0pOwogICAgfSBlbHNlIHsKICAgICAgICBjaGVja193cml0YWJpbGl0eSgpOwogICAgfQp9CgokKCcjdG9wLXNhdmUnKS5jbGljayhmdW5jdGlvbigpIHsKICAgIHZhciBrZXkgPSBnZW5lcmF0ZV9rZXkoKTsKICAgIHdyaXRlKGVuY3J5cHQoJCgnI2lucHV0JykudmFsKCksIGtleSksIGZ1bmN0aW9uKGhhc2gpIHsKICAgICAgICBpZiAoIWhhc2gpIHsKICAgICAgICAgICAgc2V0X3N0YXR1cygiRXJyb3I6IEZhaWxlZCB0byBzdG9yZSBjb250ZW50LiBJcyB0aGUgZ2F0ZXdheSB3cml0YWJsZT8iKTsKICAgICAgICB9IGVsc2UgewogICAgICAgICAgICB3aW5kb3cubG9jYXRpb24gPSAnL2lwZnMvJyArIGhhc2ggKyAnIycgKyBrZXkgKyAnLWZpcnN0dmlldyc7CiAgICAgICAgfQogICAgfSk7Cn0pOwoKJCgnI3RvcC1lZGl0JykuY2xpY2soZnVuY3Rpb24oKSB7CiAgICB1bnJlbmRlcigpOwogICAgJCgnI2lucHV0JykuZm9jdXMoKTsKICAgIGNoZWNrX3dyaXRhYmlsaXR5KCk7Cn0pOwoKJCgnI2Fib3V0JykuY2xpY2soZnVuY3Rpb24oZSkgewogICAgZS5wcmV2ZW50RGVmYXVsdCgpOwogICAgJCgnI2Fib3V0LW1vZGFsJykubW9kYWwoJ3Nob3cnKTsKICAgIGhpc3RvcnkucmVwbGFjZVN0YXRlKHVuZGVmaW5lZCwgdW5kZWZpbmVkLCAnI2Fib3V0Jyk7Cn0pOwoKJCgnI2Fib3V0LW1vZGFsJykub24oJ2hpZGUuYnMubW9kYWwnLCBmdW5jdGlvbigpIHsKICAgIGhpc3RvcnkucmVwbGFjZVN0YXRlKHVuZGVmaW5lZCwgdW5kZWZpbmVkLCAnIycgKyBkZWNyeXB0aW9uX2tleSk7Cn0pOwoKJChkb2N1bWVudCkucmVhZHkoZnVuY3Rpb24oKSB7CiAgICBsb2FkX2NvbnRlbnQoKTsKCiAgICAvLyBsb2FkIHRoZSBSRUFETUUgbW9kYWwsIGFuZCBzaG93IGl0IGlmIHRoZSBmcmFnbWVudCBpcyAjYWJvdXQKICAgICQuYWpheCh7CiAgICAgICAgdXJsOiAiUkVBRE1FLm1kIiwKICAgICAgICBzdWNjZXNzOiBmdW5jdGlvbihkYXRhKSB7CiAgICAgICAgICAgIHZhciBjID0gbmV3IHNob3dkb3duLkNvbnZlcnRlcigpOwogICAgICAgICAgICAkKCcjYWJvdXQtYm9keScpLmh0bWwoYy5tYWtlSHRtbChkYXRhKSk7CiAgICAgICAgICAgIGlmICh3aW5kb3cubG9jYXRpb24uaGFzaCA9PSAnI2Fib3V0JykgewogICAgICAgICAgICAgICAgJCgnI2Fib3V0LW1vZGFsJykubW9kYWwoJ3Nob3cnKTsKICAgICAgICAgICAgfQogICAgICAgIH0sCiAgICB9KTsKfSk7Cjwvc2NyaXB0Pgo8L2JvZHk+CjwvaHRtbD4KGKZJ"