package param

type TestValidatorReq struct {
	Name     string `json:"name" binding:"required,min=3,max=50" name:"姓名"`
	Email    string `json:"email" binding:"required,email" errMsg:"自定义错误信息:邮箱错误咯"`
	Password string `json:"password" binding:"required,min=6"`
}

type TestValidatorResp struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
