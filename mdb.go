package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type MultiDatabase struct {
	Addr string
	DbId string
}

type MultiDatabaseRequest struct {
	DbId    string        `json:"db_id"`
	SqlText string        `json:"sql_text"`
	Binds   []interface{} `json:"binds"`
}

type MultiDatabaseResult struct {
	Code   int                      `json:"code"`
	DbId   string                   `json:"db_id"`
	Error  string                   `json:"error"`
	Result []map[string]interface{} `json:"result"`
}

func (md MultiDatabase) Ping() error {
	_, err := md.Query("select * from dual")
	return err
}

func (md MultiDatabase) Query(sqlText string) ([]map[string]interface{}, error) {
	json_data, err := json.Marshal(MultiDatabaseRequest{md.DbId, sqlText, []interface{}{}})

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("http://%s/query", md.Addr)
	// resp, err := http.Post(url, "application/json",
	// 	bytes.NewBuffer(json_data))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	req.Header.Set("x-multidatabase-dbid", md.DbId)
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
