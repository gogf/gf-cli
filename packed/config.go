package packed

import "github.com/gogf/gf/os/gres"

func init() {
	if err := gres.Add("H4sIAAAAAAAC/wrwZmYRYeBg4GCoDTsXwIAE+Bk4GZLz89Iy0/VLi3L0SvJzc0JDWBkY/RifxZX0OXIdVhBwvS645/DxvO533JaeJ5Qk5i7xnbOA7dZtcV/F2C07Cv/9+T6D+0Sy9Bvmac3zV/67m9HnKewdGLfWd4/VEkvrG4YZmlbWbXFcDDceN4llnar0Nk8+Mf3Y3OUOt9hKJfxq+r0Pen9oT5dvv5rZ96Wxc6Nr3b4HltOiTzxdtOLmS4+Mi54XTeKXRV00li79u3PjQQ3NWf9kGRj+/w/wZuf42n45eRYDAwMXIwMDzH8MDBPQ/McG9x/YWwWMz+JAmpGVBHgzMokwI4IH2WBQ8MDAtkYQiSuwEKZgdwQECDD8d3wANwXJSaxsIGkmBiaGZgYGBllGEA8QAAD//0ClRg62AQAA"); err != nil {
		panic("add binary content to resource manager failed: " + err.Error())
	}
}
