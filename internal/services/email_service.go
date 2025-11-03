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
const proxyShippingNote = "195.000 VNĐ/kg" 

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

func formatJPY(amount float64) string {
    roundedAmount := int(amount + 0.5)
    s := strconv.Itoa(roundedAmount)
    n := len(s)
    if n <= 3 {
        return "¥" + s
    }

    var result strings.Builder
    result.WriteRune('¥')
    for i, r := range s {
        result.WriteRune(r)
        if (n-1-i)%3 == 0 && i != n-1 {
            result.WriteRune(',')
        }
    }
    return result.String()
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

	var currentShippingFee float64
	if order.User.VIPLevel >= 2 {
		currentShippingFee = 0 
	} else {
		currentShippingFee = float64(shippingFee)
	}
	
	finalTotal := order.TotalAmount + currentShippingFee 
	recipientEmail := os.Getenv("RECIPIENT_EMAIL")

	var itemsHTML strings.Builder
	itemsHTML.WriteString(`
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr style="background-color: #f2f2f2;">
                <th>Sản phẩm</th>
                <th>Số lượng</th>
                <th>Đơn giá</th>
                <th>Tổng</th>
            </tr>
    `)
	for _, item := range order.OrderItems {
		itemsHTML.WriteString(fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
                <td>%s</td>
                <td>%s</td>
            </tr>
        `, item.Product.Name, item.Quantity, formatVND(item.Price), formatVND(item.Price*float64(item.Quantity))))
	}
	itemsHTML.WriteString("</table>")

	body := fmt.Sprintf(`
        <h1>🎉 Bạn có đơn hàng mới!</h1>
        <p>Thông tin chi tiết đơn hàng:</p>
        %s 
        <br>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
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
    `,
		itemsHTML.String(),                    
		formatVND(order.OriginalAmount),       
		formatVND(order.OriginalAmount),        
		formatVND(order.DiscountApplied),      
		formatVND(currentShippingFee),         
		formatVND(finalTotal),                 
		order.CustomerName,                    
		order.CustomerPhone,                   
		order.CustomerEmail,                   
		order.CustomerAddress,                 
		order.PaymentMethod)                   

	subject := fmt.Sprintf("Đơn hàng mới #%d từ %s", order.ID, order.CustomerName)

	return sendEmail(recipientEmail, subject, body)
}

func SendInvoiceToCustomer(order models.Order, customerEmail string) error {
	if len(order.OrderItems) == 0 {
		return fmt.Errorf("order %d has no items", order.ID)
	}

	var currentShippingFee float64
	if order.User.VIPLevel >= 2 {
		currentShippingFee = 0 
	} else {
		currentShippingFee = float64(shippingFee)
	}

	finalTotal := order.TotalAmount + currentShippingFee
	qrImageUrl := "https://pub-be6c7e6475cd42219bb9999d8fbb5743.r2.dev/products/image.png"

	var itemsHTML strings.Builder
	itemsHTML.WriteString(`
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr style="background-color: #f2f2f2;">
                <th>Sản phẩm</th>
                <th>Số lượng</th>
                <th>Đơn giá</th>
                <th>Tổng</th>
            </tr>
    `)
	for _, item := range order.OrderItems {
		itemsHTML.WriteString(fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%d</td>
                <td>%s</td>
                <td>%s</td>
            </tr>
        `, item.Product.Name, item.Quantity, formatVND(item.Price), formatVND(item.Price*float64(item.Quantity))))
	}
	itemsHTML.WriteString("</table>")

	body := fmt.Sprintf(`
        <h1>Cảm ơn bạn đã đặt hàng tại TUNI TOKU!</h1>
        <p>Chào <b>%s</b>,</p>
        <p>Đơn hàng của bạn đã được tiếp nhận thành công. Chúng tôi sẽ sớm liên hệ với bạn để xác nhận và tiến hành giao hàng.</p>
        <h2>Chi tiết đơn hàng:</h2>
        %s
        <br>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr><td style="background-color: #f2f2f2;"><strong>Thành tiền</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Giảm giá VIP</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí ship</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Tổng thanh toán</strong></td><td><strong>%s</strong></td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin nhận hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Email</strong></td><td>%s</td></tr>
            <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>
        <h3>Mã QR code chuyển khoản:</h3>
        <p>Quét mã QR bên dưới để thanh toán(Tiền ship/Tổng tiền đơn hàng).</p>
        <img src="%s" style="width: 250px; height: 250px;" alt="QR Code">
        <p>Cảm ơn bạn đã tin tưởng và mua sắm tại TUNI TOKU!</p>
    `,
		order.CustomerName,                     
		itemsHTML.String(),                     
		formatVND(order.OriginalAmount),        
		formatVND(order.OriginalAmount),        
		formatVND(order.DiscountApplied),       
		formatVND(currentShippingFee),          
		formatVND(finalTotal),                  
		order.CustomerName,                     
		order.CustomerPhone,                    
		order.CustomerEmail,                    
		order.CustomerAddress,                 
		order.PaymentMethod,                   
		qrImageUrl)                           

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

func SendProxyOrderConfirmationEmail(order models.ProxyOrder) error {
	recipientEmail := os.Getenv("RECIPIENT_EMAIL")
	singleItemTotal := order.TotalAmountVND / float64(order.Quantity)
	basePriceVND := singleItemTotal - order.ServiceFee
	
	itemsHTML := fmt.Sprintf(`
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr style="background-color: #f2f2f2;">
                <th>Sản phẩm</th>
                <th>Giá gốc (JPY)</th>
                <th>Số lượng</th>
                <th>Tổng (VND)</th>
            </tr>
             <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%d</td>
                <td>%s</td>
            </tr>
        </table>
    `, 
    order.ProductName, 
    order.ProductName, 
    formatJPY(order.ProductPriceJPY), 
    order.Quantity, 
    formatVND(order.TotalAmountVND))

	body := fmt.Sprintf(`
        <h1>🎉 Bạn có đơn hàng đặt hộ Mercari mới!</h1>
        <p>Link gốc sản phẩm: <a href="%s">%s</a></p>
        <p>Thông tin chi tiết đơn hàng:</p>
        %s
        <br>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr><td style="background-color: #f2f2f2;"><strong>Giá quy đổi (1 sp)</strong></td><td%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí dịch vụ (1 sp)</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Thành tiền</strong></td><td><strong>%s</strong></td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí ship (dự kiến)</strong></td><td>%s (Sẽ báo sau khi hàng về kho)</td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin khách hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Email</strong></td><td>%s</td></tr>
            <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>
        <p>Vui lòng liên hệ khách hàng để xác nhận và xử lý đơn hàng.</p>
    `,
		order.MercariURL,                
		order.MercariURL,                
		itemsHTML,                       
		formatVND(basePriceVND),         
		formatVND(basePriceVND),         
		formatVND(order.ServiceFee),     
		formatVND(order.TotalAmountVND), 
		proxyShippingNote,               
		order.CustomerName,              
		order.CustomerPhone,             
		order.CustomerEmail,             
		order.CustomerAddress,           
		order.PaymentMethod)             

	subject := fmt.Sprintf("Đơn hàng đặt hộ Mercari MỚI #%d từ %s", order.ID, order.CustomerName)
	
	return sendEmail(recipientEmail, subject, body)
}

func SendProxyInvoiceToCustomer(order models.ProxyOrder, customerEmail string) error {
	qrImageUrl := "https://pub-be6c7e6475cd42219bb9999d8fbb5743.r2.dev/products/image.png"
	singleItemTotal := order.TotalAmountVND / float64(order.Quantity)
	basePriceVND := singleItemTotal - order.ServiceFee
	
	itemsHTML := fmt.Sprintf(`
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr style="background-color: #f2f2f2;">
                <th>Sản phẩm</th>
                <th>Giá gốc (JPY)</th>
                <th>Số lượng</th>
                <th>Tổng (VND)</th>
            </tr>
             <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%d</td>
                <td>%s</td>
            </tr>
        </table>
    `, 
    order.ProductName, 
    order.ProductName, 
    formatJPY(order.ProductPriceJPY), 
    order.Quantity, 
    formatVND(order.TotalAmountVND))

	body := fmt.Sprintf(`
        <h1>Cảm ơn bạn đã đặt hàng hộ tại TUNI TOKU!</h1>
        <p>Chào <b>%s</b>,</p>
        <p>Đơn hàng đặt hộ của bạn đã được tiếp nhận thành công. Chúng tôi sẽ sớm liên hệ với bạn để xác nhận và tiến hành giao hàng.</p>
        <h2>Chi tiết đơn hàng:</h2>
        %s
        <br>
        <table border="1" cellpadding="10" cellspacing="0" style="border-collapse: collapse; width: 100%;">
            <tr><td style="background-color: #f2f2f2;"><strong>Giá quy đổi (1 sp)</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí dịch vụ (1 sp)</strong></td><td>%s</td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Thành tiền</strong></td><td><strong>%s</strong></td></tr>
            <tr><td style="background-color: #f2f2f2;"><strong>Phí ship (dự kiến)</strong></td><td>%s (Sẽ báo sau khi hàng về kho)</td></tr>
            <tr><td colspan="2" style="background-color: #f2f2f2; text-align: center;"><strong>Thông tin nhận hàng</strong></td></tr>
            <tr><td><strong>Họ và tên</strong></td><td>%s</td></tr>
            <tr><td><strong>Số điện thoại</strong></td><td>%s</td></tr>
            <tr><td><strong>Email</strong></td><td>%s</td></tr>
            <tr><td><strong>Địa chỉ</strong></td><td>%s</td></tr>
            <tr><td><strong>Phương thức thanh toán</strong></td><td>%s</td></tr>
        </table>
        <>Mã QR code chuyển khoản:</    <p>Quét mã QR bên dưới để thanh toán (Tổng tiền đơn hàng). Phí ship sẽ được thanh toán khi hàng về.</p>
        <img src="%s" style="width: 250px; height: 250px;" alt="QR Code">
        <p>Cảm ơn bạn đã tin tưởng và mua sắm tại TUNI TOKU!</p>
    `,
		order.CustomerName,              
		itemsHTML,                        
		formatVND(basePriceVND),         
		formatVND(basePriceVND),         
		formatVND(order.ServiceFee),     
		formatVND(order.TotalAmountVND), 
		proxyShippingNote,
		order.CustomerName,              
		order.CustomerPhone,             
		order.CustomerEmail,             
		order.CustomerAddress,            
		order.PaymentMethod,
		qrImageUrl)

	subject := fmt.Sprintf("Xác nhận đơn hàng đặt hộ Mercari #%d từ TUNI TOKU", order.ID)

	return sendEmail(customerEmail, subject, body)
}