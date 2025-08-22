package services

import (
    "fmt"
    "os"
    "strconv"

    "github.com/kaelCoding/toyBE/internal/models"
    "gopkg.in/gomail.v2"
)

const shippingFee = 50000

func formatVND(amount float64) string {
    s := strconv.Itoa(int(amount))
    n := len(s)
    if n <= 3 {
        return s + " VNĐ"
    }

    sepCount := (n - 1) / 3
    result := make([]byte, n+sepCount)

    j := len(result) - 1
    for i := n - 1; i >= 0; i-- {
        result[j] = s[i]
        j--
        if (n-1-i)%3 == 2 && i > 0 {
            result[j] = '.'
            j--
        }
    }
    return string(result) + " VNĐ"
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

    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")
    recipientEmail := os.Getenv("RECIPIENT_EMAIL")

    body := fmt.Sprintf(`
        <h1>🎉 Bạn có đơn hàng mới!</h1>
        <p>Thông tin chi tiết đơn hàng:</p>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse;">
            <tr><td style="background-color: #f2f2f2;"><strong>Sản phẩm</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Số lượng</strong></td><td>%d</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Thành tiền</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Giảm giá VIP</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí cọc ship</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Tổng thanh toán</strong></td><td><strong>%s</strong></td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin khách hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Email</strong></td><td>%s</td></tr> <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>
        <p>Vui lòng liên hệ khách hàng để xác nhận và xử lý đơn hàng.</p>
    `, item.Product.Name, item.Quantity, formatVND(order.OriginalAmount), formatVND(order.DiscountApplied), formatVND(currentShippingFee), formatVND(finalTotal), order.CustomerName, order.CustomerPhone, order.CustomerEmail, order.CustomerAddress, order.PaymentMethod)

    m := gomail.NewMessage()
    m.SetHeader("From", smtpUser)
    m.SetHeader("To", recipientEmail)
    m.SetHeader("Subject", fmt.Sprintf("Đơn hàng mới cho sản phẩm: %s", item.Product.Name))
    m.SetBody("text/html", body)

    d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
    return d.DialAndSend(m)
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

    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")

    qrImageUrl := "https://res.cloudinary.com/dkrnoq3rb/image/upload/v1755360307/products/product_1755360303736929000.png"

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
            <tr><td style="background-color: #f2f2f2;"><strong>Phí cọc ship</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Tổng thanh toán</strong></td><td><strong>%s</strong></td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin nhận hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>

        <h3>Mã QR code chuyển khoản:</h3>
        <p>Quét mã QR bên dưới để thanh toán(Tiền cọc ship/Tổng tiền đơn hàng).</p>
        <img src="%s" style="width: 250px; height: 250px;" alt="QR Code">

        <p>Cảm ơn bạn đã tin tưởng và mua sắm tại TUNI TOKU!</p>
    `, order.CustomerName, item.Product.Name, item.Quantity, formatVND(order.OriginalAmount), formatVND(order.DiscountApplied), formatVND(currentShippingFee), formatVND(finalTotal), order.CustomerName, order.CustomerPhone, order.CustomerAddress, order.PaymentMethod, qrImageUrl)

    m := gomail.NewMessage()
    m.SetHeader("From", smtpUser)
    m.SetHeader("To", customerEmail)
    m.SetHeader("Subject", fmt.Sprintf("Xác nhận đơn hàng #%d từ TUNI TOKU", order.ID))
    m.SetBody("text/html", body)

    d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
    return d.DialAndSend(m)
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
