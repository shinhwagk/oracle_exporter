package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type MultiDatabase struct {
	addr string
	dsn  string
}

type MultiDatabaseRequest struct {
	Dsn     string        `json:"dsn"`
	SqlText string        `json:"sql_text"`
	Binds   []interface{} `json:"binds"`
}

type MultiDatabaseResult struct {
	Code   int                      `json:"code"`
	Dsn    string                   `json:"dsn"`
	Error  string                   `json:"error"`
	Result []map[string]interface{} `json:"result"`
}

func (md MultiDatabase) Ping() error {
	_, err := md.Query("select * from dual")
	return err
}

func (md MultiDatabase) Query(sqlText string) ([]map[string]interface{}, error) {
	json_data, err := json.Marshal(MultiDatabaseRequest{md.dsn, sqlText, []interface{}{}})

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s/query", md.addr)
	// resp, err := http.Post(url, "application/json",
	// 	bytes.NewBuffer(json_data))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))

	if err != nil {
		return nil, err
	}
	c := http.DefaultClient
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var mdr MultiDatabaseResult

	json.NewDecoder(resp.Body).Decode(&mdr)

	if mdr.Code == 1 {
		return nil, errors.New(mdr.Error)
	}
	return mdr.Result, nil
}
