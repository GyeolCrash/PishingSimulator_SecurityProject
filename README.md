# **FishingSimulator\_SecurityProject (피싱 시뮬레이터 백엔드)**

본 프로젝트는 LLM을 이용한 실시간 음성 기반 피싱 시뮬레이션 훈련을 위한 백엔드 서버입니다. Go(Golang), Gin, WebSocket을 기반으로 구현되었습니다.

## **1\. 기술 스택 및 환경 (Technology Stack & Environment)**

* **언어**: Go (Golang) 1.21 이상  
* **웹 프레임워크**: Gin (gin-gonic/gin)  
* **WebSocket**: gorilla/websocket  
* **인증**: golang-jwt/jwt/v5  
* **비밀번호 해싱**: golang.org/x/crypto/bcrypt  
* **동시성**: GoRoutines, Channels, sync.WaitGroup, context.Context

## **2\. 설치 및 실행 (Getting Started)**

### **2.1. 전제 조건**

* Go 1.21 이상  
* Git

### **2.2. 설치 및 실행**

1. 저장소 클론:  
```
   git clone \[https://github.com/your-username/FishingSimulator\_SecurityProject.git\](https://github.com/your-username/FishingSimulator\_SecurityProject.git)  
   cd FishingSimulator\_SecurityProject
```

2. 의존성 설치:
```  
   go mod tidy
```

3. 서버 실행:
```  
   go run cmd/api/main.go
```

   서버가 http://localhost:8080에서 실행됩니다.

### **2.3. 환경 변수 설정 (Optional)**

* 루트 디렉토리에 .env 파일을 생성하여 JWT 시크릿 키를 설정할 수 있습니다. (설정하지 않으면 internal/auth/token.go의 기본 키가 사용됩니다.)  
  JWT\_SECRET\_KEY="your\_very\_strong\_secret\_key"

### **2.4. 테스트 환경 준비 (Optional)**

* S→C (서버→클라이언트) 오디오 응답 테스트:  
  testdata/ 디렉토리를 생성하고, response.mp3 또는 response.wav 등 모의 응답으로 사용할 오디오 파일을 위치시킵니다. (파일명은 handler/websocket\_handler.go의 init() 함수에서 수정 가능)  
* C→S (클라이언트→서버) 오디오 저장 테스트:  
  서버가 실행되면 testdata/received/ 디렉토리가 자동으로 생성되며, voice 모드로 수신된 오디오 파일이 이곳에 저장됩니다.

## **3\. 디렉토리 구조 (Directory Structure)**
```
FishingSimulator\_SecurityProject/  
├── cmd/api/  
│   └── main.go               \# \[실행\] 서버 시작점, 라우터 설정  
├── internal/  
│   ├── auth/  
│   │   └── token.go          \# \[로직\] JWT 토큰 생성 및 검증  
│   ├── handler/  
│   │   ├── auth\_handler.go   \# \[핸들러\] /login, /signup HTTP 요청 처리  
│   │   └── websocket\_handler.go \# \[핸들러\] /ws/simulation WebSocket 세션 관리  
│   ├── middleware/  
│   │   └── auth.go           \# \[미들웨어\] /api/\* 경로의 JWT 인증  
│   ├── models/  
│   │   └── user.go           \# \[모델\] User 구조체 정의  
│   └── simulation/  
│       └── scenario.go       \# \[모델\] Scenario 구조체, 시나리오 데이터 정의  
├── testdata/  
│   ├── received/             \# \[테스트\] C-\>S 오디오 덤프 저장소 (자동 생성)  
│   └── response.mp3          \# \[테스트\] S-\>C 모의 응답 오디오 (수동 추가 필요)  
├── .gitignore  
├── go.mod  
├── go.sum  
└── README.md                 \# (본 파일)  
```