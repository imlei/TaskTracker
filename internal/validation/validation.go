package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"simpletask/internal/models"
)

var (
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrInvalidPhone      = errors.New("invalid phone format")
	ErrEmptyName         = errors.New("name cannot be empty")
	ErrInvalidID         = errors.New("invalid ID format")
	ErrInvalidPrice      = errors.New("price must be positive")
	ErrInvalidDate       = errors.New("invalid date format")
	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidCurrency   = errors.New("invalid currency")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordMissing   = errors.New("password must contain uppercase, lowercase, and number")
	ErrInvalidPercentage = errors.New("percentage must be between 0 and 100")
)

// Email 验证邮箱格式
func Email(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil // 允许空邮箱
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}

// Phone 验证电话号码格式（简单的国际格式验证）
func Phone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil // 允许空电话
	}
	// 移除空格、括号、破折号等常见分隔符
	cleaned := regexp.MustCompile(`[\s\-\(\)\+]`).ReplaceAllString(phone, "")
	// 验证只包含数字
	if cleaned != "" {
		matched, _ := regexp.MatchString(`^\d{7,15}$`, cleaned)
		if !matched {
			return ErrInvalidPhone
		}
	}
	return nil
}

// CustomerName 验证客户名称
func CustomerName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 100 {
		return fmt.Errorf("name too long (max 100 characters)")
	}
	return nil
}

// DisplayName 验证显示名称
func DisplayName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil // 允许空显示名称
	}
	if len(name) > 50 {
		return fmt.Errorf("display name too long (max 50 characters)")
	}
	return nil
}

// CustomerStatus 验证客户状态
func CustomerStatus(status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" || status == "active" || status == "inactive" {
		return nil
	}
	return ErrInvalidStatus
}

// Customer 验证客户数据
func Customer(c models.Customer) error {
	if err := CustomerName(c.Name); err != nil {
		return err
	}
	if err := DisplayName(c.DisplayName); err != nil {
		return err
	}
	if err := Email(c.Email); err != nil {
		return err
	}
	if err := Phone(c.Phone); err != nil {
		return err
	}
	if err := CustomerStatus(c.Status); err != nil {
		return err
	}
	// 地址验证
	if len(strings.TrimSpace(c.Address)) > 200 {
		return fmt.Errorf("address too long (max 200 characters)")
	}
	return nil
}

// TaskID 验证任务ID格式
func TaskID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidID
	}
	// 验证任务ID格式：字母开头 + 数字（如 AC0001）
	matched, _ := regexp.MatchString(`^[A-Z]+\d{4,}$`, id)
	if !matched {
		return ErrInvalidID
	}
	return nil
}

// CustomerID 验证客户ID格式
func CustomerID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrInvalidID
	}
	// 验证客户ID格式：C + 数字（如 C0001）
	matched, _ := regexp.MatchString(`^C\d{4,}$`, id)
	if !matched {
		return ErrInvalidID
	}
	return nil
}

// TaskDate 验证任务日期格式
func TaskDate(date string) error {
	date = strings.TrimSpace(date)
	if date == "" {
		return ErrInvalidDate
	}
	// 验证日期格式：YYYY-MM-DD
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, date)
	if !matched {
		return ErrInvalidDate
	}
	return nil
}

// TaskStatus 验证任务状态
func TaskStatus(status models.TaskStatus) error {
	validStatuses := map[models.TaskStatus]bool{
		models.StatusPending: true,
		models.StatusDone:    true,
		models.StatusSent:    true,
		models.StatusPaid:    true,
	}
	if !validStatuses[status] {
		return ErrInvalidStatus
	}
	return nil
}

// Price 验证价格
func Price(price *float64) error {
	if price == nil {
		return nil // 允许空价格
	}
	if *price < 0 {
		return ErrInvalidPrice
	}
	return nil
}

// Currency 验证货币代码
func Currency(c models.Currency) error {
	if c == "" {
		c = "CNY" // 默认使用人民币
	}
	c = models.Currency(strings.ToUpper(string(c)))

	validCurrencies := map[models.Currency]bool{
		"CNY": true,
		"USD": true,
		"EUR": true,
		"GBP": true,
		"JPY": true,
		"HKD": true,
	}

	if !validCurrencies[c] {
		return ErrInvalidCurrency
	}
	return nil
}

// PriceItem 验证价目项
func PriceItem(p models.PriceItem) error {
	if strings.TrimSpace(p.ServiceName) == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if len(p.ServiceName) > 100 {
		return fmt.Errorf("service name too long (max 100 characters)")
	}
	if err := Price(p.Amount); err != nil {
		return err
	}
	if err := Currency(p.Currency); err != nil {
		return err
	}
	return nil
}

// TaxRate 验证税率
func TaxRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return ErrInvalidPercentage
	}
	return nil
}

// Invoice 验证发票数据
func Invoice(inv models.Invoice) error {
	if strings.TrimSpace(inv.BillToName) == "" {
		return fmt.Errorf("bill to name cannot be empty")
	}
	if err := TaxRate(inv.TaxRate); err != nil {
		return err
	}
	if err := Currency(models.Currency(inv.Currency)); err != nil {
		return err
	}
	if inv.InvoiceDate != "" {
		if err := TaskDate(inv.InvoiceDate); err != nil {
			return err
		}
	}
	if inv.DueDate != "" {
		if err := TaskDate(inv.DueDate); err != nil {
			return err
		}
	}
	// 验证发票项
	for i, item := range inv.Items {
		if strings.TrimSpace(item.Description) == "" {
			return fmt.Errorf("invoice item %d: description cannot be empty", i)
		}
		if item.Qty <= 0 {
			return fmt.Errorf("invoice item %d: quantity must be positive", i)
		}
		if item.Rate < 0 {
			return fmt.Errorf("invoice item %d: rate cannot be negative", i)
		}
	}
	return nil
}

// Password 验证密码强度
func Password(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return ErrPasswordMissing
	}

	return nil
}

// NonEmptyString 验证非空字符串
func NonEmptyString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// MaxLength 验证字符串最大长度
func MaxLength(value string, max int, fieldName string) error {
	if len(value) > max {
		return fmt.Errorf("%s too long (max %d characters)", fieldName, max)
	}
	return nil
}

// PositiveNumber 验证正数
func PositiveNumber(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}
