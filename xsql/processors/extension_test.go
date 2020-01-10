package processors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/emqx/kuiper/common"
	"os"
	"path"
	"testing"
	"time"
)

//This cannot be run in Windows. And the plugins must be built to so before running this
//For Windows, run it in wsl with go test xsql/processors/extension_test.go xsql/processors/xsql_processor.go
func setup() *RuleProcessor {
	log := common.Log

	os.Remove(CACHE_FILE)

	dbDir, err := common.GetAndCreateDataLoc("test")
	if err != nil {
		log.Panic(err)
	}
	log.Infof("db location is %s", dbDir)

	demo := `DROP STREAM ext`
	NewStreamProcessor(demo, path.Join(dbDir, "stream")).Exec()
	demo = "CREATE STREAM ext (count bigint) WITH (DATASOURCE=\"users\", FORMAT=\"JSON\", TYPE=\"random\", CONF_KEY=\"ext\")"

	_, err = NewStreamProcessor(demo, path.Join(dbDir, "stream")).Exec()
	if err != nil{
		panic(err)
	}
	rp := NewRuleProcessor(dbDir)
	return rp
}

var CACHE_FILE = "cache"

//Test for source, sink, func and agg func extensions
//The .so files must be in the plugins folder
func TestExtensions(t *testing.T) {
	log := common.Log
	var tests = []struct {
		name    string
		rj	string
		r    [][]map[string]interface{}
	}{
		{
			name: `$$test1`,
			rj: "{\"sql\": \"SELECT echo(count) as e, countPlusOne(count) as p FROM ext where count > 49\",\"actions\": [{\"file\":  {\"path\":\"" + CACHE_FILE + "\"}}]}",

		},
	}
	fmt.Printf("The test bucket size is %d.\n\n", len(tests))
	rp := setup()
	done := make(chan struct{})
	defer close(done)
	for i, tt := range tests {
		rp.ExecDrop("$$test1")
		rs, err := rp.ExecCreate(tt.name, tt.rj)
		if err != nil {
			t.Errorf("failed to create rule: %s.", err)
			continue
		}

		tp, err := rp.ExecInitRule(rs)
		if err != nil{
			t.Errorf("fail to init rule: %v", err)
			continue
		}

		go func() {
			select {
			case err := <-tp.Open():
				log.Println(err)
				tp.Cancel()
			}
		}()
		time.Sleep(5000 * time.Millisecond)
		log.Printf("exit main program after 5 seconds")
		results := getResults()
		if len(results) == 0{
			t.Errorf("no result found")
			continue
		}
		log.Debugf("get results %v", results)
		var maps [][]map[string]interface{}
		for _, v := range results{
			var mapRes []map[string]interface{}
			err := json.Unmarshal([]byte(v), &mapRes)
			if err != nil {
				t.Errorf("Failed to parse the input into map")
				continue
			}
			maps = append(maps, mapRes)
		}

		for _, r := range maps{
			if len(r) != 1{
				t.Errorf("%d. %q\n\nresult mismatch:\n\ngot=%#v\n\n", i, tt.rj, maps)
				break
			}
			r := r[0]
			e := int((r["e"]).(float64))
			if e != 50 && e != 51{
				t.Errorf("%d. %q\n\nresult mismatch:\n\ngot=%#v\n\n", i, tt.rj, maps)
				break
			}
			p := int(r["p"].(float64))
			if p != 2 {
				t.Errorf("%d. %q\n\nresult mismatch:\n\ngot=%#v\n\n", i, tt.rj, maps)
				break
			}
		}
		tp.Cancel()
	}
}

func getResults() []string{
	f, err := os.Open(CACHE_FILE)
	if err != nil{
		panic(err)
	}
	defer f.Close()
	result := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan(){
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return result
}
