package model

import "time"

// UserRealm represents a user's trust realm
type UserRealm struct {
	DirectTrust []TrustingUser `json:"direct_trust"`
}

// TrustingUser is the base for all trust related verifications
type TrustingUser struct {
	Id         string    `json:"-"`
	SetId      int       `json:"set_id"`
	UserId     string    `json:"user_id"`
	StartTrust time.Time `json:"start"`
	EndTrust   time.Time `json:"end"`
}

type CreateUserUserTrust struct {
	TrustedUserId string    `json:"trusted_user_id"`
	TrustStart    time.Time `json:"trust_start"`
	TrustEnd      time.Time `json:"trust_end"`
}

type DeleteUserUserTrust struct {
	SetId int `json:"set_id"`
}

type SigningRequest struct {
	Message                []byte    `json:"message"`
	ServerAppliedTimestamp time.Time `json:"-"`
	UserId                 string    `json:"-"`
}

type SigningResponse struct {
	Signature  []byte `json:"signature"`
	CommitHash string `json:"commit_hash"`
}

type VerificationRequest struct {
	Message   []byte `json:"message"`
	Signature []byte `json:"signature"`
}

type VerificationResponse struct {
	Verified bool   `json:"verified"`
	ErrorMsg string `json:"error_msg"`
}

type LoginRequest struct {
	AuthMethod string `json:"auth_method"`
	AuthToken  string `json:"auth_token"`
}

type LoginResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type RegisterRequest struct {
	Password string `json:"password"`
}

type RegisterResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type TokenHolder interface {
	GetToken() string
	GetUsername() string
}

func (l LoginResponse) GetUsername() string {
	return l.Username
}

func (r RegisterResponse) GetUsername() string {
	return r.Username
}

func (l LoginResponse) GetToken() string {
	return l.Token
}

func (r RegisterResponse) GetToken() string {
	return r.Token
}
