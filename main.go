package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	DB *gorm.DB
)

type Todo struct {
	ID     int       `json:"id"`
	Info   string    `json:"info"`
	Date   time.Time `json:"date"`
	Status bool      `json:"status"`
}

func initMySQL() (err error) {
	dsn := "root:030223@tcp(127.0.0.1:3306)/todolist?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open("mysql", dsn)
	if err != nil {
		return
	}
	return DB.DB().Ping()
}

func main() {

	//获取当前时间
	nowDate := time.Now().Format("2006-01-02T15:04:05Z")

	//连接数据库
	err := initMySQL()
	if err != nil {
		panic(err)
	}
	defer DB.Close()
	//模型绑定
	DB.AutoMigrate(&Todo{})

	r := gin.Default()
	r.Static("/assets", "assets")
	r.LoadHTMLGlob("templates/*")
	//跨域
	r.Use(cors.Default())
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	editGroup := r.Group("edit")
	{

		//添加
		editGroup.POST("/todo", func(c *gin.Context) {
			//前端页面填写待办事项，点击提交，发请求
			//从请求中把数据拿出来
			var todo Todo
			c.BindJSON(&todo)
			//更新nowDate
			nowDate = todo.Date.Format("2006-01-02T15:04:05Z")

			//去掉时间偏移
			todo.Date = todo.Date.Add(-8 * time.Hour)
			//存入数据库//返回响应
			err = DB.Debug().Create(&todo).Error
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, todo)
			}

		})
		//查看
		//查看所有代办事项
		editGroup.GET("/todo/bydate/:date", func(c *gin.Context) {
			//接收时间信息
			selectedDate, _ := c.Params.Get("date")
			startTime, _ := time.Parse("2006-01-02", selectedDate)
			endTime, _ := time.Parse("2006-01-02", selectedDate)
			endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
			//更新当前日期
			nowDate = startTime.Format("2006-01-02T15:04:05Z")
			//查询表里的所有数据
			var todoList []Todo
			if err = DB.Debug().Where("`date` BETWEEN ? AND ?", startTime, endTime).Find(&todoList).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, todoList)
			}
		})
		//得到当前日期
		editGroup.GET("/todo/getdate", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"date": nowDate})
		})
		//查看某一个代办事项
		editGroup.GET("/todo/byid/:id", func(c *gin.Context) {

		})
		//修改
		editGroup.PUT("/todo/byid/:id", func(c *gin.Context) {
			id, ok := c.Params.Get("id")
			if !ok {
				c.JSON(http.StatusOK, gin.H{"error": "无效id"})
				return
			}
			var todo Todo
			if err = DB.Where("`id`=?", id).First(&todo).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				return
			}
			c.BindJSON(&todo)
			todo.Status = !todo.Status
			DB.Model(&Todo{}).Where("`id`=?", id).Update("`status`", &todo.Status)
			if err = DB.Save(&todo).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, todo)
			}

		})
		//删除
		editGroup.DELETE("/todo/byid/:id", func(c *gin.Context) {
			id, ok := c.Params.Get("id")
			if !ok {
				c.JSON(http.StatusOK, gin.H{"error": "无效id"})
				return
			}
			if err = DB.Where("`id`=?", id).Delete(Todo{}).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{id: "deleted success！"})
			}
		})
	}

	r.Run()
}
