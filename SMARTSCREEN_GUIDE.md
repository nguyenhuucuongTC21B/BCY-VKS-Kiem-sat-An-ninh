# 🛡️ Hướng dẫn xử lý SmartScreen Warning

> **Vấn đề:** Mỗi lần chạy `BCY-VKS.exe`, Windows hiện popup:
> 
> 🚫 **"Windows protected your PC"** — Microsoft Defender SmartScreen prevented an unrecognized app from starting.
> 
> Phải bấm **"More info"** → **"Run anyway"** mới chạy được.

Đây là cơ chế bảo mật mặc định của Windows. Bài toán là file `.exe` chưa được **code-signed** (ký số) với chứng chỉ từ nhà cung cấp chứng thực (CA).

---

## 🎯 4 giải pháp (từ đơn giản → chuyên nghiệp)

### 🟢 Cách 1: Unblock file (1 lần, miễn phí, chỉ máy đó)

**Khi nào dùng:** Tải `BCY-VKS.exe` về máy, chạy 1 lần, popup xuất hiện → muốn bỏ popup vĩnh viễn trên máy đó.

```powershell
# Chạy PowerShell với quyền Admin
powershell -ExecutionPolicy Bypass -File unblock-file.ps1 -ExePath "C:\path\to\BCY-VKS.exe"

# Hoặc bằng tay:
# 1. Right-click BCY-VKS.exe → Properties
# 2. Tick checkbox "Unblock" ở tab General
# 3. Apply → OK
# 4. Chạy .exe bình thường - không còn SmartScreen
```

**Tác dụng:** Windows gỡ "Mark of the Web" (đánh dấu tải từ Internet) khỏi file → SmartScreen không còn cảnh báo.

**Hạn chế:** Chỉ tác dụng trên máy đó, với file đó. Nếu copy sang máy khác hoặc tải lại, popup quay lại.

---

### 🟡 Cách 2: Self-signed cert (miễn phí, trust máy đã cài cert)

**Khi nào dùng:** Sử dụng nội bộ trong cơ quan (máy cùng domain, có Group Policy phân phối cert).

```powershell
# Chạy PowerShell với quyền Admin
powershell -ExecutionPolicy Bypass -File sign-code.ps1 selfsigned
```

**Quá trình script tự động:**
1. Tạo self-signed code signing cert trong `Cert:\CurrentUser\My`
2. Copy cert vào `Trusted Root CA` và `Trusted Publishers` của máy
3. Export backup ra `build/BCY-VKS-code-signing.pfx` (password: `BCY-VKS-2026`)
4. Gọi `signtool.exe sign` để ký số file `BCY-VKS.exe`
5. Verify signature

**Kết quả:**
- ✅ Trên máy đã chạy script: SmartScreen **không xuất hiện** nữa
- ❌ Trên máy khác: vẫn cảnh báo cho đến khi cài cert đó vào `Trusted Root` + `Trusted Publishers`

**Phân phối cert cho máy khác:**

```powershell
# Trên máy đã ký, export cert ra file .cer:
$cert = Get-ChildItem Cert:\CurrentUser\My | Where-Object {$_.Subject -match "BCY-VKS"}
Export-Certificate -Cert $cert -FilePath "BCY-VKS-cert.cer"

# Copy BCY-VKS-cert.cer sang máy khác, chạy:
# Cài vào Trusted Root:
Import-Certificate -FilePath "BCY-VKS-cert.cer" -CertStoreLocation Cert:\LocalMachine\Root
# Cài vào Trusted Publishers:
Import-Certificate -FilePath "BCY-VKS-cert.cer" -CertStoreLocation Cert:\LocalMachine\TrustedPublisher
```

**Hoặc qua Group Policy** (cho domain doanh nghiệp):
1. Mở **Group Policy Management** → tạo GPO
2. Computer Configuration → Policies → Windows Settings → Security Settings → Public Key Policies → **Trusted Publishers** và **Trusted Root Certification Authorities**
3. Import cert từ file `.cer`
4. Link GPO đến OU cần áp dụng
5. Trên máy client: `gpupdate /force`

---

### 🟠 Cách 3: OV Code Signing cert ($200-300/năm)

**Khi nào dùng:** Phần mềm thương mại, hoặc dùng trong nhiều cơ quan khác nhau, không thể tự cài cert từng máy.

**Mua cert từ CA:**
- **DigiCert** Code Signing: https://www.digicert.com/code-signing/
- **Sectigo** (Comodo) Code Signing: https://sectigo.com/ssl-certificates/code-signing
- **GlobalSign** Code Signing: https://www.globalsign.com/en/code-signing-certificate/

**Quy trình:**
1. Đăng ký với thông tin tổ chức (cần giấy phép kinh doanh, mã số thuế)
2. CA xác minh (KYC) → mất 1-7 ngày
3. Nhận file `.pfx` (chứa private key + cert)
4. Build .exe bình thường
5. Ký số:
   ```powershell
   powershell -ExecutionPolicy Bypass -File sign-code.ps1 production `
       -PfxPath "C:\certs\BCY-VKS-production.pfx" `
       -CertPassword "your_pfx_password"
   ```

**Hạn chế:** OV cert vẫn cần **reputation building** trên SmartCloud của Microsoft. Trong tuần đầu, có thể vẫn có popup SmartScreen ở một số máy. Sau khi có reputation (khi nhiều người đã chạy), popup sẽ biến mất.

---

### 🔴 Cách 4: EV Code Signing cert ($400-700/năm, immediately trusted)

**Khi nào dùng:** Phần mềm thương mại, muốn ngay lập tức không còn popup SmartScreen.

**Khác biệt với OV:**
- Yêu cầu xác minh mở rộng (Extended Validation) với hardware token USB
- **Immediately trusted** — không cần reputation building
- **Bypass SmartScreen hoàn toàn** trên mọi máy từ lần đầu tiên

**Quy trình:**
1. Đăng ký với CA
2. CA gửi USB token (Secure Key Storage) chứa private key
3. Build .exe
4. Cắm USB token, ký số:
   ```powershell
   powershell -ExecutionPolicy Bypass -File sign-code.ps1 production `
       -PfxPath "C:\certs\BCY-VKS-EV.pfx" `
       -CertPassword "your_usb_token_pin"
   ```

---

## 📋 Bảng so sánh 4 cách

| Cách | Chi phí | Phạm vi áp dụng | Bypass SmartScreen | Khó khăn |
|------|---------|------------------|--------------------|---------|
| **1. Unblock** | Miễn phí | 1 file, 1 máy | ✅ 1 lần | Phải làm từng máy |
| **2. Self-signed** | Miễn phí | Nhiều máy (cần cài cert) | ✅ Sau khi cài cert | Phải phân phối cert |
| **3. OV cert** | $200-300/năm | Toàn cầu | ⚠️ Tuần đầu có thể cảnh báo | Cần thời gian reputation |
| **4. EV cert** | $400-700/năm | Toàn cầu | ✅ Ngay lập tức | Cần USB token |

---

## 🚀 Khuyến nghị cho Ban Cơ yếu

### **Quy trình triển khai nội bộ (miễn phí):**

1. **Trên máy dev build:**
   ```powershell
   # Build + ký số với self-signed cert
   .\build.bat
   # Hoặc ký riêng:
   powershell -ExecutionPolicy Bypass -File sign-code.ps1 selfsigned
   ```

2. **Export cert ra file `.cer`:**
   ```powershell
   $cert = Get-ChildItem Cert:\CurrentUser\My | Where-Object {$_.Subject -match "BCY-VKS"}
   Export-Certificate -Cert $cert -FilePath "\\fileserver\BCY-VKS-cert.cer"
   ```

3. **Phân phối cert qua Group Policy:**
   - Computer Configuration → Windows Settings → Security Settings → Public Key Policies
   - Import vào **Trusted Root Certification Authorities** + **Trusted Publishers**
   - Link GPO đến OU toàn cơ quan

4. **Phân phối .exe:**
   - Copy `BCY-VKS.exe` + `BCY-VKS-cert.cer` qua file share
   - Trên máy client: chạy `gpupdate /force`, sau đó chạy `.exe` → không còn popup

### **Quy trình production (mua cert):**

1. **Mua OV cert từ DigiCert** (~$289/năm) — https://www.digicert.com/code-signing/
2. **Đăng ký tên tổ chức:** "Ban Co Yeu Trung Uong"
3. **Nhận .pfx** sau khi xác minh
4. **Ký số:**
   ```powershell
   powershell -ExecutionPolicy Bypass -File sign-code.ps1 production `
       -PfxPath "C:\certs\BCY-VKS-OV.pfx" `
       -CertPassword "secure_password"
   ```
5. **Build reputation:** Trong 1-2 tuần đầu, chạy .exe trên nhiều máy có telemetry → reputation tăng → popup giảm

---

## 📞 Hỗ trợ kỹ thuật

- **SmartScreen:** https://learn.microsoft.com/windows/security/threat-protection/windows-defender-smartscreen/
- **Code signing certs:** https://learn.microsoft.com/windows-hardware/drivers/install/authenticode
- **Group Policy deployment:** https://learn.microsoft.com/windows-server/identity/ad-ds/enterprise-security/group-policy-overview

---

*Lưu ý pháp lý: Ký số không có nghĩa là file sạch virus. Ký số chỉ chứng minh nguồn gốc tổ chức phát hành. Người dùng vẫn cần kiểm tra tính hợp lệ của tổ chức phát hành.*
