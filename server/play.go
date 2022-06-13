package server

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)


type hand uint

const (
	NilHand   hand = iota //空白
	BlackHand             //黑手
	WhiteHand             //白手
	Win       int = 5
)

func (h hand) Str() string {
	switch h {
	case NilHand:
		return "."
	case BlackHand:
		return "X"
	case WhiteHand:
		return "O"
	default:
		return "."
	}
}

type Grid struct {
	value  hand
	left   *Grid
	right  *Grid
	top    *Grid
	bottom *Grid
}
//左上
func (g *Grid) LeftTop() *Grid {
	if g.left != nil {
		return g.left.top
	}
	return nil
}
//左下
func (g *Grid) LeftBottom() *Grid {
	if g.left != nil {
		return g.left.bottom
	}
	return nil
}
//右上
func (g *Grid) RightTop() *Grid {
	if g.right != nil {
		return g.right.top
	}
	return nil
}
//右下
func (g *Grid) RightBottom() *Grid {
	if g.right != nil {
		return g.right.bottom
	}
	return nil
}

//设置row,col坐标的值, 即 落棋子
func (g *Grid) Set(row, col int, value hand) {
	offset := g
	offset = g.Offset(row, col)
	if offset == nil {
		return
	} else if offset.value == NilHand {
		offset.value = value
	}
	return
}

//获取向右偏移x位的指针（未使用）
func (g *Grid) RightOffset(col int) *Grid {
	var tmp *Grid
	tmp = g
	if g == nil {
		return nil
	}
	for i := 1; i <= col; i++ {
		if i == col && tmp != nil {
			return tmp
		}
		if tmp == nil {
			return nil
		}
		if tmp.right != nil {
			tmp = tmp.right
		}
	}
	return nil
}

//获取向下偏移y位的指针（未使用）
func (g *Grid) BottomOffset(row int) *Grid {
	var tmp *Grid
	tmp = g
	if g == nil {
		return nil
	}
	for i := 1; i <= row; i++ {
		if i == row && tmp != nil {
			return tmp
		}
		if tmp == nil {
			return nil
		}
		if tmp.bottom != nil {
			tmp = tmp.bottom
		}
	}
	return nil
}

//获取该表格有几行
func (g *Grid) GetRowLen() int {
	var row int
	tmp := g
	for tmp != nil {
		row++
		tmp = tmp.bottom
	}
	return row
}

//获取该表格有几列
func (g *Grid) GetColLen() int {
	var col int
	tmp := g
	for tmp != nil {
		col++
		tmp = tmp.right
	}
	return col
}
//打印棋盘布局
func (g *Grid) Print() {
	fmt.Println("当前棋盘布局为:")
	var colNumStr = ""
	col := g.GetColLen()
	fillC := strconv.Itoa(g.GetColLen())
	for i := 1; i <= col; i++ {
		colNumStr += " " + StrLeftFill(len(fillC), i)
	}
	fmt.Println(StrLeftFill(len(strconv.Itoa(g.GetRowLen())), ""), strings.TrimLeft(colNumStr, " "))

	for row := 1; row <= g.GetRowLen(); row++ {
		var rowStr = ""
		for col := 1; col <= g.GetColLen(); col++ {
			//rowStr += StrLeftFill(len(strconv.Itoa(g.GetColLen())), "") + g.Offset(row, col).value.Str()
			rowStr += " " + StrLeftFill(len(strconv.Itoa(g.GetColLen())), g.Offset(row, col).value.Str())
		}
		fmt.Println(StrLeftFill(len(strconv.Itoa(g.GetRowLen())), row), strings.TrimLeft(rowStr, " "))
	}

	fmt.Println("")
}

/*
获取坐标处的指针
*/
func (g *Grid) Offset(row, col int) *Grid {
	offset := g
	if g == nil {
		return nil
	}
	var i, j int
	i, j = 0, 0
	for i <= row {
		if row == i {
			for j <= col {
				if col == j {
					return offset
				}

				if offset == nil {
					return nil
				}
				offset = offset.right
				j++
			}
		}
		if offset == nil {
			return nil
		}
		offset = offset.bottom
		i++
	}
	return nil
}

//棋盘是否已满？
func (g *Grid) IsFull() bool {
	for row := 0; row < g.GetRowLen(); row++ {
		for col := 0; col < g.GetColLen(); col++ {
			if g.Offset(row, col).value == NilHand {
				return false
			}
		}
	}
	return true
}

//统计黑手,白手的棋子数量
func (g *Grid) Count() (black, white int) {
	for row := 0; row < g.GetRowLen(); row++ {
		for col := 0; col < g.GetColLen(); col++ {
			if g.Offset(row, col).value == BlackHand {
				black++
			} else if g.Offset(row, col).value == WhiteHand {
				white++
			}
		}
	}
	return
}

//检查是否已分出胜负
func (g *Grid) IsWin(row, col int) int {
	//获取落子点的坐标
	offset := g.Offset(row, col)
	h := offset.value
	//没下
	if h == NilHand {
		return 0
	}
	//检查行
	count := 1
	left := offset.left
	right := offset.right
	//检测左边
	for {
		if left == nil {
			break
		}
		if left.value == h {
			count++
		} else {
			break
		}
		left = left.left
	}
	//检测右边
	for {
		if right == nil {
			break
		}
		if right.value == h {
			count++
		} else {
			break
		}
		right = right.right
	}
	if count >= Win {
		switch h {
		case WhiteHand:
			fmt.Println("白手赢-左右", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return  2
		case BlackHand:
			fmt.Println("黑手赢-左右", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return  1
		}
		return 0
	}

	//检查列
	count = 1
	top := offset.top
	bottom := offset.bottom
	//上
	for {
		if top == nil {
			break
		}
		if top.value == h {
			count++
		} else {
			break
		}
		top = top.top
	}
	//下
	for {
		if bottom == nil {
			break
		}
		if bottom.value == h {
			count++
		} else {
			break
		}
		bottom = bottom.bottom
	}
	if count >= Win {
		switch h {
		case WhiteHand:
			fmt.Println("白手赢-上下", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return 2
		case BlackHand:
			fmt.Println("黑手赢-上下", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return 1
		}
		return 0
	}

	//检查左斜边
	count = 1
	leftTop := offset.LeftTop()
	rightBottom := offset.RightBottom()
	//左上
	for {
		if leftTop == nil {
			break
		}
		if leftTop.value == h {
			count++
		} else {
			break
		}
		leftTop = leftTop.LeftTop()
	}
	//右下
	for {
		if rightBottom == nil {
			break
		}
		if rightBottom.value == h {
			count++
		} else {
			break
		}
		rightBottom = rightBottom.RightBottom()
	}
	if count >= Win {
		switch h {
		case WhiteHand:
			fmt.Println("白手赢-左斜边", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return 2
		case BlackHand:
			fmt.Println("黑手赢-左斜边", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return 1
		}
		return 0
	}

	//检查右斜边
	count = 1
	rightTop := offset.RightTop()
	leftBottom := offset.LeftBottom()
	//右上
	for {
		if rightTop == nil {
			break
		}
		if rightTop.value == h {
			count++
		} else {
			break
		}
		rightTop = rightTop.RightTop()
	}
	//左下
	for {
		if leftBottom == nil {
			break
		}
		if leftBottom.value == h {
			count++
		} else {
			break
		}
		leftBottom = leftBottom.LeftBottom()
	}
	if count >= Win {
		switch h {
		case WhiteHand:
			fmt.Println("白手赢-右斜边", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return 2
		case BlackHand:
			fmt.Println("黑手赢-右斜边", fmt.Sprintf("最后一个落子点为(row:%d,col:%d)", row, col))
			return 1
		}
		return 0
	}
	if g.IsFull() {
		fmt.Println("平手")
		return 3
	} else {
		return 0
	}
}

/*
约定: row,col的起始值为1
*/
func InitGrid(row, col int, head *Grid) *Grid {
	l := &Grid{}

	l = head
	c := col
	for c > 0 {
		var tmp Grid
		tmp.left = l
		l.right = &tmp
		l = &tmp
		c--
	}

	l = head
	r := row
	for r > 0 {
		var tmp Grid
		tmp.top = l
		l.bottom = &tmp
		l = &tmp
		r--
	}

	for i := 1; i <= row; i++ {
		for j := 1; j <= col; j++ {
			top := head.Offset(i-1, j)
			left := head.Offset(i, j-1)
			var tmp Grid
			tmp.top = top
			tmp.left = left

			top.bottom = &tmp
			left.right = &tmp
		}
	}
	return head
}

func TestGrid(row, col int) {
	/*
		简单的测试下棋子
		设置:
				(1,4),(4,3)为黑手
				(3,4),(4,6)为白手
	*/
	grid := InitGrid(row, col, &Grid{})

	grid.Set(1, 4, BlackHand)
	grid.Set(4, 3, BlackHand)
	grid.Set(3, 4, WhiteHand)
	grid.Set(4, 6, WhiteHand)
	grid.Print()
}

//落子的坐标
type XY struct {
	row int
	col int
}

func GameTwo(row, col int) {
	/*
		游戏规则:(和我们平时玩的规则一样)
			1. 一行连续5个子
			2. 一列连续5个子
			3. 对角线连续5个子
			4. 棋盘被完全填满,还未出现胜负,则平局
	*/
	grid := InitGrid(row, col, &Grid{})

	//生成所有的棋子位置
	xy := map[int]XY{}
	var loop int // 第几次循环
	for r := 1; r <= row; r++ {
		for c := 1; c <= col; c++ {
			xy[loop] = XY{row: r, col: c}
			loop++
		}
	}
	//fmt.Println(xy)
	rand.Seed(time.Now().Unix())
	p := XY{1, 1}
	i := 0
	for {
		if grid.IsWin(p.row, p.col) > 0 {
			grid.Print()
			break
		}

		//随机落棋
		for {
			if v, ok := xy[rand.Intn(loop+1)]; ok {
				p = v
				delete(xy,loop) //已取出,删除该坐标
				break
			}
			//棋子坐标非法， continue
		}

		if i%2 == 0 {
			//黑手落棋子
			grid.Set(p.row, p.col, BlackHand)
		} else {
			//白手落棋子
			grid.Set(p.row, p.col, WhiteHand)
		}
		i++
	}
}

//字符串左边填充
func StrLeftFill(s int, value interface{}) string {
	var format = ""
	format = "%" + strconv.Itoa(s) + "v"
	return fmt.Sprintf(format, value)
}

//func main() {
//	grid := InitGrid(6, 7, &Grid{})
//
//	//空棋盘
//	grid.Print()
//
//	TestGrid(6, 7)
//	GameTwo(10, 10)
//}


