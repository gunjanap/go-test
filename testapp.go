package microapp

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Used
	log "github.com/sirupsen/logrus"
)

// TestApp Provides convinience methods for test
type TestApp struct {
	application             *App
	controllerRouteProvider func(*App) []RouteSpecifier
	dbInitializer           func(db *gorm.DB)
}

// NewTestApp returns new instance of TestApp
func NewTestApp(appName string, controllerRouteProvider func(*App) []RouteSpecifier, dbInitializer func(db *gorm.DB), verbose bool) *TestApp {
	dbFile := "./test_islax.db"
	db, err := gorm.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}

	db.LogMode(verbose)

	logger := log.New()
	application := New(appName, nil, logger, db, nil)

	return &TestApp{application: application, controllerRouteProvider: controllerRouteProvider, dbInitializer: dbInitializer}
}

// Initialize prepares the app for testing
func (testApp *TestApp) Initialize() {
	testApp.application.Initialize(testApp.controllerRouteProvider(testApp.application))
	testApp.PrepareEmptyTables()

	go testApp.application.Start()
}

// Stop the app
func (testApp *TestApp) Stop() {
	testApp.application.Stop()
	testApp.application.DB.Close()
}

// PrepareEmptyTables clears all table of data
func (testApp *TestApp) PrepareEmptyTables() {
	testApp.dbInitializer(testApp.application.DB)
}

// ExecuteRequest executes the http request
func (testApp *TestApp) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	testApp.application.Router.ServeHTTP(rr, req)

	return rr
}

// CheckResponseCode checks if the http response is as expected
func (testApp *TestApp) CheckResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

// GetToken gets a token to connect to API
func (testApp *TestApp) GetToken(tenantID string, userID string, scope []string) string {
	return testApp.generateToken(tenantID, userID, "", "", uuid.UUID{}.String(), "", scope, false)
}

// GetAdminToken returns a test token
func (testApp *TestApp) GetAdminToken(tenantID string, userID string, scope []string) string {
	return testApp.generateToken(tenantID, userID, "", "", uuid.UUID{}.String(), "", scope, true)
}

// GetFullAdminToken returns a test token with all the fields along with different external IDs for types such as Appliance, Session, User. These external IDs are used with REST api is invoked from another REST API service as opposed to the getting hit from UI by the user.
func (testApp *TestApp) GetFullAdminToken(tenantID string, userID string, username string, name string, externalID string, externalIDType string, scope []string) string {
	return testApp.generateToken(tenantID, userID, username, name, externalID, externalIDType, scope, true)
}

// GetFullToken returns a test token with all the fields along with different external IDs for types such as Appliance, Session, User. These external IDs are used with REST api is invoked from another REST API service as opposed to the getting hit from UI by the user.
func (testApp *TestApp) GetFullToken(tenantID string, userID string, username string, name string, externalID string, externalIDType string, scope []string) string {
	return testApp.generateToken(tenantID, userID, username, name, externalID, externalIDType, scope, false)
}

// generateToken generates and return token
func (testApp *TestApp) generateToken(tenantID string, userID string, username string, name string, externalID string, externalIDType string, scope []string, admin bool) string {
	hmacSampleSecret := []byte(testApp.application.Config.GetString("ISLA_JWT_SECRET"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":              "http://isla.cyberinc.com",
		"aud":              "http://isla.cyberinc.com",
		"iat":              time.Now().Unix(),
		"exp":              time.Now().Add(time.Minute * 60).Unix(), // Expires in 1 hour
		"tenant":           tenantID,
		"user":             userID,
		"admin":            admin,
		"name":             username,
		"displayName":      name,
		"scope":            scope,
		"externalId":       externalID,
		"externalIdType":   externalIDType,
		"identityProvider": "",
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSampleSecret)

	if err != nil {
		panic(err)
	}

	return tokenString
}
