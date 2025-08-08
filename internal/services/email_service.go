package services

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kaelCoding/toyBE/internal/models"
	"gopkg.in/gomail.v2"
)

func SendOrderConfirmationEmail(order models.Order) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	recipientEmail := os.Getenv("RECIPIENT_EMAIL")

	// Soạn nội dung email bằng HTML để đẹp hơn
	body := fmt.Sprintf(`
		<h1>🎉 Bạn có đơn hàng mới!</h1>
		<p>Thông tin chi tiết đơn hàng:</p>
		<table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse;">
			<tr><td style="background-color: #f2f2f2;"><strong>Sản phẩm</strong></td><td>%s</td></tr>
			<tr><td style="background-color: #f2f2f2;"><strong>Số lượng</strong></td><td>%d</td></tr>
			<tr><td style="background-color: #f2f2f2;"><strong>Tổng tiền</strong></td><td>%d VND</td></tr>
			<tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin khách hàng</strong></td></tr>
			<tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
			<tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
			<tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
			<tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
		</table>
		<p>Vui lòng liên hệ khách hàng để xác nhận và xử lý đơn hàng.</p>
	`, order.ProductName, order.Quantity, order.TotalPrice, order.CustomerName, order.CustomerPhone, order.CustomerAddress, order.PaymentMethod)

	m := gomail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", fmt.Sprintf("Đơn hàng mới cho sản phẩm: %s", order.ProductName))
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	// Gửi email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func SendFeedbackEmail(feedback models.Feedback) error {
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")
    recipientEmail := os.Getenv("RECIPIENT_EMAIL")

    body := fmt.Sprintf(`
        <h1>💡 Bạn có một góp ý mới từ người dùng!</h1>
        <p>Thông tin chi tiết góp ý:</p>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse;">
            <tr><td style="background-color: #f2f2f2;"><strong>Tên người gửi</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Email</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Nội dung góp ý</strong></td><td>%s</td></tr>
        </table>
        <p>Vui lòng xem xét góp ý này để cải thiện dịch vụ.</p>
    `, feedback.Name, feedback.Email, feedback.Content)

    m := gomail.NewMessage()
    m.SetHeader("From", smtpUser)
    m.SetHeader("To", recipientEmail)
    m.SetHeader("Subject", fmt.Sprintf("Góp ý mới từ: %s", feedback.Name))
    m.SetBody("text/html", body)

    d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

    if err := d.DialAndSend(m); err != nil {
        return err
    }

    return nil
}