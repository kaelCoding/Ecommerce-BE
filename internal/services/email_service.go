package services

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"

    "github.com/kaelCoding/toyBE/internal/models"
    "github.com/resend/resend-go/v2"
)

const shippingFee = 50000

func formatVND(amount float64) string {
    roundedAmount := int(amount + 0.5)
    s := strconv.Itoa(roundedAmount)
    n := len(s)
    if n <= 3 {
        return s + " VNĐ"
    }

    var result strings.Builder
    for i, r := range s {
        result.WriteRune(r)
        if (n-1-i)%3 == 0 && i != n-1 {
            result.WriteRune('.')
        }
    }
    return result.String() + " VNĐ"
}

func sendEmail(to, subject, htmlBody string) error {
    apiKey := os.Getenv("RESEND_API_KEY")
    fromEmail := os.Getenv("RESEND_FROM_EMAIL")

    if apiKey == "" || fromEmail == "" {
        return fmt.Errorf("RESEND_API_KEY and RESEND_FROM_EMAIL must be set")
    }

    client := resend.NewClient(apiKey)

    params := &resend.SendEmailRequest{
        From:    fromEmail,
        To:      []string{to},
        Subject: subject,
        Html:    htmlBody,
    }

    sent, err := client.Emails.Send(params)
    if err != nil {
        return fmt.Errorf("error sending email: %w", err)
    }

    log.Printf("Email sent successfully to %s, ID: %s\n", to, sent.Id)
    return nil
}

func SendOrderConfirmationEmail(order models.Order) error {
    if len(order.OrderItems) == 0 {
        return fmt.Errorf("order %d has no items", order.ID)
    }
    item := order.OrderItems[0]

    currentShippingFee := float64(shippingFee)
    if order.User.VIPLevel >= 2 {
        currentShippingFee = 0
    }
    finalTotal := order.TotalAmount + currentShippingFee
    recipientEmail := os.Getenv("RECIPIENT_EMAIL")

    body := fmt.Sprintf(`
        <h1>🎉 Bạn có đơn hàng mới!</h1>
        <p>Thông tin chi tiết đơn hàng:</p>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse;">
            <tr><td style="background-color: #f2f2f2;"><strong>Sản phẩm</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Số lượng</strong></td><td>%d</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Thành tiền</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Giảm giá VIP</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí ship</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Tổng thanh toán</strong></td><td><strong>%s</strong></td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin khách hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Email</strong></td><td>%s</td></tr>
            <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>
        <p>Vui lòng liên hệ khách hàng để xác nhận và xử lý đơn hàng.</p>
    `, item.Product.Name, item.Quantity, formatVND(order.OriginalAmount), formatVND(order.DiscountApplied), formatVND(currentShippingFee), formatVND(finalTotal), order.CustomerName, order.CustomerPhone, order.CustomerEmail, order.CustomerAddress, order.PaymentMethod)

    subject := fmt.Sprintf("Đơn hàng mới cho sản phẩm: %s", item.Product.Name)
    
    return sendEmail(recipientEmail, subject, body)
}

func SendInvoiceToCustomer(order models.Order, customerEmail string) error {
    if len(order.OrderItems) == 0 {
        return fmt.Errorf("order %d has no items", order.ID)
    }
    item := order.OrderItems[0]

    currentShippingFee := float64(shippingFee)
    if order.User.VIPLevel >= 2 {
        currentShippingFee = 0
    }
    finalTotal := order.TotalAmount + currentShippingFee

    qrImageUrl := "https://res.cloudinary.com/dkrnoq3rb/image/upload/v1757779793/Screenshot_2025-09-13_at_11.09.34_PM_oyt8cn.png"

    body := fmt.Sprintf(`
        <h1>Cảm ơn bạn đã đặt hàng tại TUNI TOKU!</h1>
        <p>Chào <b>%s</b>,</p>
        <p>Đơn hàng của bạn đã được tiếp nhận thành công. Chúng tôi sẽ sớm liên hệ với bạn để xác nhận và tiến hành giao hàng.</p>
        <h2>Chi tiết đơn hàng:</h2>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse;">
            <tr><td style="background-color: #f2f2f2;"><strong>Sản phẩm</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Số lượng</strong></td><td>%d</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Thành tiền</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Giảm giá VIP</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí ship</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Tổng thanh toán</strong></td><td><strong>%s</strong></td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin nhận hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>
        <h3>Mã QR code chuyển khoản:</h3>
        <p>Quét mã QR bên dưới để thanh toán(Tiền ship/Tổng tiền đơn hàng).</p>
        <img src="%s" style="width: 250px; height: 250px;" alt="QR Code">
        <p>Cảm ơn bạn đã tin tưởng và mua sắm tại TUNI TOKU!</p>
    `, order.CustomerName, item.Product.Name, item.Quantity, formatVND(order.OriginalAmount), formatVND(order.DiscountApplied), formatVND(currentShippingFee), formatVND(finalTotal), order.CustomerName, order.CustomerPhone, order.CustomerAddress, order.PaymentMethod, qrImageUrl)

    subject := fmt.Sprintf("Xác nhận đơn hàng #%d từ TUNI TOKU", order.ID)

    return sendEmail(customerEmail, subject, body)
}

func SendFeedbackEmail(feedback models.Feedback) error {
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

    subject := fmt.Sprintf("Góp ý mới từ: %s", feedback.Name)
    
    return sendEmail(recipientEmail, subject, body)
}