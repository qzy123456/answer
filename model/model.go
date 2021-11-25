package model

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	Db *gorm.DB
)

func NewModel(url, user, pass, port, db string) {
	var err error
	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, url, port, db)
	Db, err = gorm.Open("mysql", dbUrl)
	if err != nil {
		fmt.Println("数据库链接失败")
		panic(err)
	}
}

type Users struct {
	UserId   int
	UserName string
}

// 题库信息表
type KsQuestion struct {
	Id              int    `gorm:"column:id" db:"id" json:"id" form:"id"`                                                         //题目ID
	QuestionTitle   string `gorm:"column:question_title" db:"question_title" json:"question_title" form:"question_title"`         //题目问题
	QuestionContent string `gorm:"column:question_content" db:"question_content" json:"question_content" form:"question_content"` //题目说明
	QuestionImg     string `json:"question_img" gorm:"column:question_img"`                                                       // 问题图片
	AnswerType      int8   `gorm:"column:answer_type" db:"answer_type" json:"answer_type" form:"answer_type"`                     //答案类型(1 单选 2 判断 3 多选 )
	Answers         string `gorm:"column:answers" db:"answers" json:"answers" form:"answers"`                                     //问题信息
	AnswerId        int    `gorm:"column:answer_id" db:"answer_id" json:"answer_id" form:"answer_id"`                             //正确答案ID
}

/**
 * 查询所有的题目
 */
func GetAllExamId() (map[int]KsQuestion, error) {
	var Exam []KsQuestion
	res := make(map[int]KsQuestion, 0)
	Db.Table("ks_question").Where("special_id = 4 and status = 1").Order("RAND()").Limit(1).Find(&Exam)
	for key, value := range Exam {
		res[key] = value
	}
	return res, nil
}
