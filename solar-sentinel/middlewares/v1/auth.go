package v1
import (
	"net/http"
	"strings"
	"github.com/negeek/solar-sphere/solar-sentinel/utils"
	model"github.com/negeek/solar-sphere/solar-sentinel/repository/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
		)

func AuthenticationMiddleware(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			accessKey string
			err error
			accessKeyArr []string
			bearer string
			key string
			claim = &utils.UserClaim{} 
			user = &shared.User{}
			exist bool
		)
		accessKey = r.Header.Get("Authorization")
		if accessKey==""{
			utils.JsonResponse(w, false, http.StatusUnauthorized , "Provide access key", nil)
			return	
		}
		// verify the jwt token
		accessKeyArr = strings.Split(accessKey, " ")

		if len(accessKeyArr) != 2 {
			utils.JsonResponse(w, false, http.StatusUnauthorized, "Invalid Authorisation Header", nil)
			return
		}
		bearer, key = accessKeyArr[0], accessKeyArr[1]
		if bearer!="Bearer"{
			utils.JsonResponse(w, false, http.StatusUnauthorized , "Invalid Authorisation Header", nil)
			return	
		}
		
		claim, err = utils.VerifyAccessKey(key)
		if err != nil{
			utils.JsonResponse(w, false, http.StatusBadRequest ,"Invalid access key", nil)
			return	
		}

		// verify claims 
		user.Email=claim.Email
		exist = model.FindUserByEmail(user)
		if exist != true {
			utils.JsonResponse(w, false, http.StatusUnauthorized ,"Invalid User", nil)	
			return
		}
        handler.ServeHTTP(w, r)
    })
}