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
