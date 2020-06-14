package boot

import "github.com/gogf/gf/os/gres"

func init() {
	if err := gres.Add("H4sIAAAAAAAC/wrwZmYRYeBg4GDQ8zsXwIAE+Bk4GZLz89Iy0/VLi3L0SvJzc0JDWBkYV3x6GlfSd7Cr2UDE9TrbJ/7pZ8rUbvL/UFVKSv605apNp+Ct/fqdrLszTr7bU833k7Fyzd69B4s+CD88NMdneqf0rubMpD+tRROV9iv17rHICS7a4HwhyKSMY60Iw4bjkqZXZlza9Vfa2U5/nQ3bJbMViwteVc/ZPO+g+PzaN+LLLS95qx2T+vmv4fmM/056T10S05Ode926VlVpT+2cd2PF/X5x6Z5Ji7LmMzIw/P8f4M3OoVP/vXsxAwODBCMDA8yjDAzBaB5lg3sU7L/3n57GgTQjKwnwZmQSYUaEE7LBoHCCgW2NIBJXqCFMwe4ICBBg+O/4Em4KkpNY2UDSTAxMDM0MDAxqjCAeIAAA//+dGqFMvwEAAA=="); err != nil {
		panic("add binary content to resource manager failed: " + err.Error())
	}
}
