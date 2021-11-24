package model

import (
	"fmt"
	"os"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	mondb  *mgo.Session
	currdb *mgo.Database

	UserTable = "Users"
)

type Model struct {
}

func NewModel(url string) {
	m := new(Model)
	m.connect(url)
}

func (this *Model) connect(url string) {
	mondb, err := mgo.Dial(url)
	if err != nil {
		fmt.Println("Mongo Connect Error: ", err)
		os.Exit(-1)
	}

	//defer mondb.Close()

	mondb.SetMode(mgo.Monotonic, true)
	currdb = mondb.DB("Vckai")
}

type Status struct {
	Id_       bson.ObjectId `bson:"_id"`
	UserIndex int
	ExamIndex int
}

type Users struct {
	UserId      int
	UserName    string
}

type Exam struct {
	Id_          bson.ObjectId `bson:"_id"`
	ExamId       int
	ExamQuestion string
	ExamOption   []string
	ExamAnwser   int
	ExamResolve  string
	ExamTime     time.Time
}

/**
 * 获取状态
 */
func GetStatus() Status {
	c := currdb.C("status")
	status := Status{}
	err := c.Find(nil).One(&status)
	if err != nil {
		Id_ := bson.NewObjectId()
		c.Insert(&Status{Id_: Id_, UserIndex: 0, ExamIndex: 0})
		status.Id_ = Id_
		status.UserIndex = 0
	}
	return status
}

/**
 * add exam..
 */
func AddExam(question string, options []string, anwser int, resolve string) (int, error) {
	c := currdb.C("exam")
	status := GetStatus()

	index := status.ExamIndex + 1

	err := c.Insert(&Exam{
		Id_:          bson.NewObjectId(),
		ExamId:       index,
		ExamQuestion: question,
		ExamOption:   options,
		ExamAnwser:   anwser,
		ExamResolve:  resolve,
		ExamTime:     time.Now(),
	})
	if err == nil { //题目ID累加
		c2 := currdb.C("status")
		c2.Update(nil, bson.M{"$inc": bson.M{"examindex": 1}})
	}

	return index, err
}

/**
 * get all exam ids..
 */
func GetAllExamId() ([]int, error) {
	c := currdb.C("exam")
	var result []Exam
	err := c.Find(nil).Select(bson.M{"examid": 1}).All(&result)
	if err != nil {
		return []int{}, err
	}
	var ret []int
	for _, v := range result {
		ret = append(ret, v.ExamId)
	}
	fmt.Println(result)
	return ret, nil
}

/**
 * get exam by examid.
 */
func GetExam(examId int) (Exam, error) {
	c := currdb.C("exam")
	var result Exam
	err := c.Find(bson.M{"examid": examId}).One(&result)
	return result, err
}
