package model

import (
	"database/sql"
	"github.com/golang/glog"
	"time"
)

var tableName = "interview"
var completeTable = "complete_list"

type IssueStruct struct {
	Id            int
	Issue         string
	Answer        string
	RelatedIssues string
	Knowledge     string
	Tips          string
}

// TODO 用泛型实现查询不同字段时的兼容
// TODO 看下别的sql库是怎么实现兼容查询不同字段的
func GetIssueIds(priority int, issueType string) []int {
	var condition1, condition2 string
	condition1 = "priority > ?"
	if priority > 0 {
		condition1 = "priority = ?"
	}
	if issueType != "" {
		condition2 = " and type = ?"
	}
	sqlStr := "select id from " + tableName + " where " + condition1 + condition2 + " and is_read = 0"
	//fmt.Println(sqlStr)
	//os.Exit(0)
	stmt, err := master.Prepare(sqlStr)
	if err != nil {
		glog.Errorln("prepare failed,sql:", sqlStr, ",err:", err)
		return nil
	}
	defer stmt.Close()
	var rows *sql.Rows
	if issueType != "" {
		rows, err = stmt.Query(priority, issueType)
	} else {
		rows, err = stmt.Query(priority)
	}
	if err != nil {
		glog.Errorln("query failed,sql:", sqlStr, ", err:%v\n", err)
		return nil
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			glog.Errorln("查询mysql失败，err:", err)
			return nil
		}
		//TODO 这里循环append会不会有性能问题，有没更好的方案 => 看看其他sql库怎么做的
		ids = append(ids, id)
	}
	return ids
}

func GetIssueById(id int) (*IssueStruct, error) {
	issue := &IssueStruct{}
	//获取issue
	sqlStr1 := "select id,issue,answer,related_issues,tips,knowledge from " + tableName + " where id = ?"
	stmt, queryErr := master.Prepare(sqlStr1)
	if queryErr != nil {
		glog.Errorln("exec sql1 failed, err:", queryErr)
		return nil, queryErr
	}
	defer stmt.Close()
	err := stmt.QueryRow(id).Scan(&issue.Id, &issue.Issue, &issue.Answer, &issue.RelatedIssues, &issue.Tips, &issue.Knowledge)
	if err != nil {
		glog.Errorln("query failed,sql:", sqlStr1, ", err:%v\n", err)
		return nil, err
	}

	//更新该issue为已读
	sqlStr2 := "Update " + tableName + " set is_read=1 where id=?"
	stmt, UpdateErr := master.Prepare(sqlStr2)
	if UpdateErr != nil {
		glog.Errorln("exec sql2 failed, err:", UpdateErr)
		return nil, UpdateErr
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		glog.Errorln("update failed, err:", err)
		return nil, err
	}

	return issue, nil
}

func GetCompleteList(id int) ([]string, error) {
	sqlStr := "select create_at from " + completeTable + " where interview_id = ? order by create_at desc limit 3"
	stmt, queryErr := master.Prepare(sqlStr)
	if queryErr != nil {
		glog.Errorln("exec sql1 failed, err:", queryErr)
		return nil, queryErr
	}
	defer stmt.Close()
	rows, err := stmt.Query(id)
	if err != nil {
		glog.Errorln("query failed,sql:", sqlStr, ", err:%v\n", err)
		return nil, err
	}
	defer rows.Close()
	createAtList := make([]string, 0)
	for rows.Next() {
		var createAt string
		if err := rows.Scan(&createAt); err != nil {
			glog.Errorln("scan failed,err:", err)
			return createAtList, err
		}
		outputTime, err := time.Parse(time.RFC3339, createAt)
		if err != nil {
			glog.Errorln("Parse time failed:", err)
			return createAtList, err
		}
		//TODO 这里循环append会不会有性能问题，有没更好的方案 => 看看其他sql库怎么做的
		createAtList = append(createAtList, outputTime.Format("2006-01-02 15:04:05"))
	}
	return createAtList, nil
}

// 更新所有is_read为0
func Reset() {
	//更新该issue为已读
	sqlStr := "Update " + tableName + " set is_read=0"
	stmt, UpdateErr := master.Prepare(sqlStr)
	if UpdateErr != nil {
		glog.Errorln("exec sql failed, err:", UpdateErr)
		return
	}
	defer stmt.Close()
	_, err := stmt.Exec()
	if err != nil {
		glog.Errorln("update failed, err:", err)
		return
	}

	return
}

func Complete(id string) error {
	tx, err := master.Begin()
	if err != nil {
		return err
	}

	// 准备更新表 interview
	updateStmt, err := tx.Prepare("UPDATE interview SET last_complete = ? WHERE id = ?")
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}
	defer updateStmt.Close()

	// 执行更新语句
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = updateStmt.Exec(now, id)
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}

	// 准备插入一条记录到表 complete_lisr 中
	insertStmt, err := tx.Prepare("INSERT INTO complete_list (interview_id) VALUES (?)")
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}
	defer insertStmt.Close()

	// 执行插入语句
	_, err = insertStmt.Exec(id)
	if err != nil {
		// 回滚事务
		tx.Rollback()
		return err
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
