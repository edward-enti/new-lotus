package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type seedCategory struct {
	Name, Slug, Short, Desc, Surface, Icon string
	Sort                                   int
}

type seedProduct struct {
	CategorySlug, SKU, Name, Slug, Role, Finish, Usage, Summary, Desc, Guide string
	Features                                                                 []string
	Specs                                                                    map[string]any
	Packaging                                                                []string
	Colors                                                                   []map[string]string
	Coverage                                                                 float64
	Featured                                                                 bool
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

var seedCategories = []seedCategory{
	{"Sơn Gỗ Hệ Nước", "son-go-he-nuoc", "Giải pháp sơn gỗ nội ngoại thất cao cấp, đạt chuẩn xuất khẩu Châu Âu/Mỹ.", "Sơn gỗ hệ nước Sơn Lotus không độc hại, giữ vân gỗ tự nhiên hoặc phủ màu hiện đại, phù hợp đồ gỗ nội thất và ngoại thất.", "wood", "paint-roller", 1},
	{"Sơn Giả Gỗ", "son-gia-go", "Sơn tạo vân gỗ trên sắt mạ kẽm, bê tông, tấm xi măng sợi, nhôm đúc.", "Sơn giả gỗ tạo vân gỗ chân thực trên bề mặt không phải gỗ: sắt mạ kẽm, bê tông, Smartboard/Cemboard, nhôm đúc.", "faux_wood", "paint-roller", 2},
	{"Sơn Kim Loại Hệ Nước", "son-kim-loai-he-nuoc", "Sơn chống rỉ và sơn phủ trực tiếp lên sắt mạ kẽm, inox, nhôm, đồng.", "Độ bám dính cực cao, chống rỉ sét, thi công trực tiếp không cần sơn lót riêng cho nhiều bề mặt kim loại.", "metal", "paint-roller", 3},
	{"Sơn Ngói & Bê Tông", "son-ngoi-be-tong", "Sơn ngói, sơn giả đá, sơn sân tennis, sàn bê tông chống thấm.", "Kháng kiềm, chống rêu mốc, chịu thời tiết cho mái ngói, sân và sàn bê tông.", "concrete_tile", "paint-roller", 4},
	{"Sơn Loa (Speaker Paint)", "son-loa", "Sơn thùng loa chuyên dụng tạo gai sần, chống trầy xước, chống thấm.", "Bề mặt gai sần (texture) bền cho loa sân khấu và hội trường, chống trầy xước và chống thấm nước.", "speaker", "paint-roller", 5},
	{"Sơn Composite & Nhựa", "son-composite-nhua", "Sơn chuyên dụng cho vật liệu FRP, nhựa ABS, PVC, chậu kiểng composite.", "Bám dính tốt trên composite và nhựa, màng sơn dẻo dai, không bong tróc khi bề mặt co giãn nhẹ.", "composite", "paint-roller", 6},
	{"Trám Trét & Chống Thấm", "tram-tret-chong-tham", "Keo bóng nước, keo chống thấm, bột trét gỗ hệ nước.", "Trám trét và bảo vệ bề mặt gỗ, tường, bê tông khỏi thấm nước, nứt chân chim và ẩm mốc.", "putty_waterproof", "paint-roller", 7},
}

var seedProducts = []seedProduct{
	{
		CategorySlug: "son-go-he-nuoc", SKU: "SL-ACR-001",
		Name: "Sơn Phủ Gỗ Hệ Nước Lotus Acrylic Lacquer", Slug: "son-phu-go-he-nuoc-lotus-acrylic-lacquer",
		Role: "topcoat", Finish: "gloss", Usage: "both",
		Summary:   "Sơn phủ gỗ hệ nước cao cấp, màng bóng đẹp, giữ vân gỗ tự nhiên.",
		Desc:      "Sơn phủ Acrylic hệ nước cho đồ gỗ nội ngoại thất. Màng sơn trong, bóng, chống ố vàng, không mùi khó chịu khi thi công.",
		Guide:     "Làm sạch và chà nhám bề mặt. Sơn lót 1-2 lớp, chà nhám nhẹ, sau đó sơn phủ 2 lớp, mỗi lớp cách nhau 2 giờ.",
		Features:  []string{"Hệ nước 100% không mùi", "Giữ vân gỗ tự nhiên", "Chống ố vàng", "Kháng tia UV"},
		Specs:     map[string]any{"coverage_m2_per_liter": 10.0, "recommended_coats": 2, "dry_time_touch_mins": 30, "dry_time_recoat_hours": 2, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "LW-01", "name": "Cánh Gián", "hex": "#6E2C00", "swatch_url": ""}, {"code": "LW-02", "name": "Nâu Đậm", "hex": "#4A2C1A", "swatch_url": ""}},
		Coverage:  10.0, Featured: true,
	},
	{
		CategorySlug: "son-go-he-nuoc", SKU: "SL-PRM-001",
		Name: "Sơn Lót Gỗ Hệ Nước Lotus Wood Primer", Slug: "son-lot-go-he-nuoc-lotus-wood-primer",
		Role: "primer", Finish: "matte", Usage: "both",
		Summary:   "Sơn lót gỗ hệ nước, tăng bám dính và che phủ cho lớp phủ.",
		Desc:      "Sơn lót chuyên dụng cho gỗ, lấp đầy thớ gỗ, tăng độ bám dính cho sơn phủ và tiết kiệm sơn phủ.",
		Guide:     "Chà nhám gỗ, lau sạch bụi. Lăn hoặc phun 1-2 lớp sơn lót, chà nhám nhẹ trước khi sơn phủ.",
		Features:  []string{"Hệ nước 100% không mùi", "Tăng bám dính", "Lấp thớ gỗ"},
		Specs:     map[string]any{"coverage_m2_per_liter": 9.0, "recommended_coats": 2, "dry_time_touch_mins": 30, "dry_time_recoat_hours": 2, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  9.0, Featured: false,
	},
	{
		CategorySlug: "tram-tret-chong-tham", SKU: "SL-FIL-001",
		Name: "Bột Trét Gỗ Hệ Nước Lotus Wood Filler", Slug: "bot-tret-go-he-nuoc-lotus-wood-filler",
		Role: "putty", Finish: "matte", Usage: "both",
		Summary:   "Bột trét gỗ hệ nước, trám lỗ đinh, vết nứt và khuyết tật bề mặt gỗ.",
		Desc:      "Bột trét dẻo, dễ chà nhám, không co ngót nhiều, có thể sơn phủ trực tiếp sau khi khô.",
		Guide:     "Trét đầy vết khuyết, để khô hoàn toàn, chà nhám phẳng rồi sơn lót và sơn phủ.",
		Features:  []string{"Hệ nước 100% không mùi", "Dễ chà nhám", "Không co ngót nhiều"},
		Specs:     map[string]any{"coverage_m2_per_liter": 8.0, "recommended_coats": 1, "dry_time_touch_mins": 60, "dry_time_recoat_hours": 4, "dilution": "Không pha loãng"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L"},
		Colors:    []map[string]string{},
		Coverage:  8.0, Featured: false,
	},
	{
		CategorySlug: "son-gia-go", SKU: "SL-FWG-001",
		Name: "Sơn Giả Gỗ Lotus Trên Sắt Mạ Kẽm & Xi Măng", Slug: "son-gia-go-lotus-sat-ma-kem-xi-mang",
		Role: "topcoat", Finish: "satin", Usage: "both",
		Summary:   "Sơn tạo vân gỗ chân thực trên sắt mạ kẽm và xi măng, bền ngoài trời.",
		Desc:      "Hệ sơn giả gỗ cho cửa cổng, hàng rào, lan can sắt mạ kẽm và tấm xi măng. Vân gỗ vẽ tay kết hợp lớp phủ bảo vệ chống tia UV và mưa nắng.",
		Guide:     "Vệ sinh bề mặt, sơn lót chống rỉ nếu cần. Tạo nền, vẽ vân gỗ, phủ bóng bảo vệ 2 lớp.",
		Features:  []string{"Hệ nước 100% không mùi", "Vân gỗ chân thực", "Kháng tia UV", "Chống rỉ sét"},
		Specs:     map[string]any{"coverage_m2_per_liter": 8.0, "recommended_coats": 2, "dry_time_touch_mins": 45, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "FW-01", "name": "Vân Gỗ Sồi", "hex": "#B07A3F", "swatch_url": ""}, {"code": "FW-02", "name": "Vân Gỗ Óc Chó", "hex": "#5C3A21", "swatch_url": ""}},
		Coverage:  8.0, Featured: true,
	},
	{
		CategorySlug: "son-kim-loai-he-nuoc", SKU: "SL-MET-001",
		Name: "Sơn Kim Loại Mạ Kẽm Hệ Nước Lotus 2-in-1", Slug: "son-kim-loai-ma-kem-he-nuoc-lotus",
		Role: "topcoat", Finish: "gloss", Usage: "both",
		Summary:   "Sơn 2-in-1 chống rỉ và phủ màu trực tiếp trên sắt mạ kẽm, inox, nhôm.",
		Desc:      "Không cần sơn lót chống rỉ riêng cho nhiều bề mặt kim loại. Bám dính cực cao, màng cứng, chống rỉ sét và chịu thời tiết.",
		Guide:     "Vệ sinh, tẩy dầu mỡ và rỉ lỏng. Sơn trực tiếp 2 lớp, mỗi lớp cách nhau 2-3 giờ.",
		Features:  []string{"Hệ nước 100% không mùi", "Chống rỉ sét", "Bám dính cực cao", "2-in-1 lót và phủ"},
		Specs:     map[string]any{"coverage_m2_per_liter": 9.0, "recommended_coats": 2, "dry_time_touch_mins": 40, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "MT-01", "name": "Đen Mờ", "hex": "#1F2937", "swatch_url": ""}, {"code": "MT-02", "name": "Trắng Sứ", "hex": "#F8FAFC", "swatch_url": ""}},
		Coverage:  9.0, Featured: true,
	},
	{
		CategorySlug: "son-loa", SKU: "SL-SPK-001",
		Name: "Sơn Loa Tạo Gai Sần Lotus Speaker Paint", Slug: "son-loa-tao-gai-san-lotus-speaker-paint",
		Role: "topcoat", Finish: "matte", Usage: "both",
		Summary:   "Sơn thùng loa tạo gai sần, chống trầy xước và chống thấm nước.",
		Desc:      "Bề mặt gai sần đều, cứng, chịu va đập cho thùng loa sân khấu và hội trường. Chống thấm nước, dễ lau chùi.",
		Guide:     "Vệ sinh thùng loa, sơn lót nếu bề mặt hút sơn mạnh. Phun tạo gai bằng súng chuyên dụng, phủ 2 lớp.",
		Features:  []string{"Hệ nước 100% không mùi", "Tạo gai sần đều", "Chống trầy xước", "Chống thấm nước"},
		Specs:     map[string]any{"coverage_m2_per_liter": 6.0, "recommended_coats": 2, "dry_time_touch_mins": 60, "dry_time_recoat_hours": 4, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "SP-01", "name": "Đen Gai", "hex": "#111827", "swatch_url": ""}},
		Coverage:  6.0, Featured: false,
	},
	{
		CategorySlug: "son-ngoi-be-tong", SKU: "SL-ROF-001",
		Name: "Sơn Ngói Cao Cấp Hệ Nước Lotus Roof Paint", Slug: "son-ngoi-cao-cap-he-nuoc-lotus",
		Role: "topcoat", Finish: "satin", Usage: "outdoor",
		Summary:   "Sơn ngói hệ nước kháng kiềm, chống rêu mốc, bền màu ngoài trời.",
		Desc:      "Cho mái ngói, sân bê tông và tường ngoại thất. Kháng kiềm, chống rêu mốc, chống thấm nhẹ và chịu mưa nắng.",
		Guide:     "Rửa sạch rêu mốc và bụi bẩn, để khô. Sơn lót kháng kiềm 1 lớp, sơn phủ 2 lớp.",
		Features:  []string{"Hệ nước 100% không mùi", "Kháng kiềm", "Chống rêu mốc", "Chịu thời tiết"},
		Specs:     map[string]any{"coverage_m2_per_liter": 7.0, "recommended_coats": 2, "dry_time_touch_mins": 60, "dry_time_recoat_hours": 4, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "RF-01", "name": "Đỏ Ngói", "hex": "#9A3412", "swatch_url": ""}, {"code": "RF-02", "name": "Xám Đá", "hex": "#475569", "swatch_url": ""}},
		Coverage:  7.0, Featured: true,
	},
	{
		CategorySlug: "tram-tret-chong-tham", SKU: "SL-SEA-001",
		Name: "Keo Bóng Nước Chống Thấm Lotus Clear Sealer", Slug: "keo-bong-nuoc-chong-tham-lotus-clear-sealer",
		Role: "sealer", Finish: "gloss", Usage: "both",
		Summary:   "Keo bóng nước trong suốt, chống thấm và bảo vệ bề mặt gỗ, đá, bê tông.",
		Desc:      "Lớp phủ trong suốt, không đổi màu bề mặt, chống thấm nước, chống bám bẩn và tăng độ bóng nhẹ.",
		Guide:     "Bề mặt khô sạch. Quét hoặc phun 2 lớp mỏng, mỗi lớp cách nhau 2-3 giờ.",
		Features:  []string{"Hệ nước 100% không mùi", "Trong suốt", "Chống thấm nước", "Không đổi màu bề mặt"},
		Specs:     map[string]any{"coverage_m2_per_liter": 12.0, "recommended_coats": 2, "dry_time_touch_mins": 30, "dry_time_recoat_hours": 2, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  12.0, Featured: false,
	},
}

var seedProductsExtra = []seedProduct{
	{
		CategorySlug: "son-gia-go", SKU: "SL-FWG-002",
		Name: "Sơn Lót Giả Gỗ Lotus Wood Base", Slug: "son-lot-gia-go-lotus-wood-base",
		Role: "primer", Finish: "matte", Usage: "both",
		Summary:   "Sơn lót tạo nền cho hệ giả gỗ, tăng bám dính và đều màu nền.",
		Desc:      "Lớp lót cho hệ sơn giả gỗ trên sắt mạ kẽm, bê tông và tấm xi măng. Nền đều giúp vân gỗ lên rõ và lớp phủ bám chắc hơn.",
		Guide:     "Vệ sinh bề mặt, xử lý rỉ lỏng và dầu mỡ. Sơn 1-2 lớp lót, để khô hoàn toàn rồi mới tạo vân gỗ.",
		Features:  []string{"Hệ nước 100% không mùi", "Tăng bám dính", "Nền đều màu"},
		Specs:     map[string]any{"coverage_m2_per_liter": 9.0, "recommended_coats": 2, "dry_time_touch_mins": 40, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  9.0, Featured: false,
	},
	{
		CategorySlug: "son-gia-go", SKU: "SL-FWG-003",
		Name: "Sơn Phủ Bóng Giả Gỗ Lotus Clear Top", Slug: "son-phu-bong-gia-go-lotus-clear-top",
		Role: "topcoat", Finish: "gloss", Usage: "both",
		Summary:   "Lớp phủ bóng trong suốt bảo vệ vân gỗ khỏi mưa nắng và tia UV.",
		Desc:      "Lớp phủ bảo vệ cho bề mặt giả gỗ đã tạo vân. Màng trong, bóng nhẹ, chống thấm và giữ vân gỗ không bạc màu nhanh dưới nắng.",
		Guide:     "Chỉ phủ khi lớp vân đã khô hoàn toàn. Quét hoặc phun 2 lớp mỏng, mỗi lớp cách nhau 2-3 giờ.",
		Features:  []string{"Hệ nước 100% không mùi", "Trong suốt", "Kháng tia UV", "Chống thấm nước"},
		Specs:     map[string]any{"coverage_m2_per_liter": 11.0, "recommended_coats": 2, "dry_time_touch_mins": 30, "dry_time_recoat_hours": 2, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  11.0, Featured: true,
	},
	{
		CategorySlug: "son-kim-loai-he-nuoc", SKU: "SL-MET-002",
		Name: "Sơn Lót Chống Rỉ Lotus Rust Primer", Slug: "son-lot-chong-ri-lotus-rust-primer",
		Role: "primer", Finish: "matte", Usage: "both",
		Summary:   "Sơn lót chống rỉ cho sắt thép, sắt mạ kẽm trước khi phủ màu.",
		Desc:      "Lớp lót ức chế rỉ sét cho khung sắt, cửa cổng và kết cấu kim loại. Bám dính tốt trên bề mặt đã làm sạch rỉ lỏng.",
		Guide:     "Chà sạch rỉ lỏng, tẩy dầu mỡ, để khô. Sơn 1-2 lớp lót, sau đó sơn phủ màu khi lớp lót đã khô.",
		Features:  []string{"Hệ nước 100% không mùi", "Chống rỉ sét", "Tăng bám dính"},
		Specs:     map[string]any{"coverage_m2_per_liter": 8.0, "recommended_coats": 2, "dry_time_touch_mins": 45, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  8.0, Featured: false,
	},
	{
		CategorySlug: "son-kim-loai-he-nuoc", SKU: "SL-MET-003",
		Name: "Sơn Phủ Kim Loại Mờ Lotus Metal Matte", Slug: "son-phu-kim-loai-mo-lotus-metal-matte",
		Role: "topcoat", Finish: "matte", Usage: "both",
		Summary:   "Sơn phủ kim loại bề mặt mờ, hiện đại, ít lộ vết xước nhỏ.",
		Desc:      "Sơn phủ mờ cho lan can, khung cửa và đồ kim loại nội ngoại thất. Bề mặt mờ giúp che khuyết tật nhỏ và ít chói dưới nắng.",
		Guide:     "Làm sạch bề mặt, sơn lót chống rỉ nếu cần. Sơn phủ 2 lớp, khuấy đều trước và trong khi thi công.",
		Features:  []string{"Hệ nước 100% không mùi", "Bề mặt mờ", "Chống rỉ sét", "Chịu thời tiết"},
		Specs:     map[string]any{"coverage_m2_per_liter": 9.0, "recommended_coats": 2, "dry_time_touch_mins": 40, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "MT-03", "name": "Xám Chì", "hex": "#334155", "swatch_url": ""}, {"code": "MT-04", "name": "Xanh Rêu", "hex": "#365314", "swatch_url": ""}},
		Coverage:  9.0, Featured: false,
	},
	{
		CategorySlug: "son-ngoi-be-tong", SKU: "SL-ROF-002",
		Name: "Sơn Lót Kháng Kiềm Lotus Alkali Primer", Slug: "son-lot-khang-kiem-lotus-alkali-primer",
		Role: "primer", Finish: "matte", Usage: "outdoor",
		Summary:   "Sơn lót kháng kiềm cho tường, ngói và bê tông mới.",
		Desc:      "Lớp lót cho bề mặt xi măng, bê tông và ngói mới, nơi kiềm còn cao. Giúp lớp phủ không bị phấn hóa sớm và bám chắc hơn.",
		Guide:     "Bề mặt phải khô và sạch bụi. Sơn 1 lớp lót kháng kiềm, để khô hoàn toàn rồi sơn phủ.",
		Features:  []string{"Hệ nước 100% không mùi", "Kháng kiềm", "Tăng bám dính"},
		Specs:     map[string]any{"coverage_m2_per_liter": 8.0, "recommended_coats": 1, "dry_time_touch_mins": 60, "dry_time_recoat_hours": 4, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  8.0, Featured: false,
	},
	{
		CategorySlug: "son-ngoi-be-tong", SKU: "SL-ROF-003",
		Name: "Sơn Sàn Bê Tông Lotus Floor Coat", Slug: "son-san-be-tong-lotus-floor-coat",
		Role: "topcoat", Finish: "satin", Usage: "outdoor",
		Summary:   "Sơn sàn bê tông chịu mài mòn cho sân, lối đi và sàn ngoài trời.",
		Desc:      "Sơn sàn cho bê tông sân vườn, lối đi và khu vực ngoài trời. Màng dai, chống trơn trượt nhẹ nhờ bề mặt sần mịn và chịu rửa trôi.",
		Guide:     "Rửa sạch bụi, dầu mỡ và rêu mốc, để khô. Sơn lót kháng kiềm 1 lớp, sơn sàn 2 lớp, tránh đi lại trong 24 giờ đầu.",
		Features:  []string{"Hệ nước 100% không mùi", "Chịu mài mòn", "Chống rêu mốc", "Chịu thời tiết"},
		Specs:     map[string]any{"coverage_m2_per_liter": 6.0, "recommended_coats": 2, "dry_time_touch_mins": 90, "dry_time_recoat_hours": 6, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "FL-01", "name": "Xám Bê Tông", "hex": "#64748B", "swatch_url": ""}},
		Coverage:  6.0, Featured: true,
	},
	{
		CategorySlug: "son-loa", SKU: "SL-SPK-002",
		Name: "Sơn Lót Thùng Loa Lotus Speaker Primer", Slug: "son-lot-thung-loa-lotus-speaker-primer",
		Role: "primer", Finish: "matte", Usage: "both",
		Summary:   "Sơn lót cho thùng loa gỗ và MDF trước khi tạo gai sần.",
		Desc:      "Lớp lót làm đều bề mặt thùng loa, giảm hút sơn và giúp lớp gai sần bám đều, không loang lổ.",
		Guide:     "Trám kín khe hở, chà nhám và lau bụi. Sơn lót 1-2 lớp, chà nhám nhẹ rồi mới phun tạo gai.",
		Features:  []string{"Hệ nước 100% không mùi", "Làm đều bề mặt", "Tăng bám dính"},
		Specs:     map[string]any{"coverage_m2_per_liter": 7.0, "recommended_coats": 2, "dry_time_touch_mins": 60, "dry_time_recoat_hours": 4, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  7.0, Featured: false,
	},
	{
		CategorySlug: "son-loa", SKU: "SL-SPK-003",
		Name: "Sơn Loa Gai Mịn Lotus Fine Texture", Slug: "son-loa-gai-min-lotus-fine-texture",
		Role: "topcoat", Finish: "matte", Usage: "both",
		Summary:   "Sơn loa gai mịn cho thùng loa cần bề mặt tinh tế, ít thô ráp.",
		Desc:      "Bề mặt gai mịn, đều tay cho thùng loa trong nhà và phòng thu nhỏ. Màng cứng, chống trầy xước nhẹ và dễ lau bụi.",
		Guide:     "Sơn lót trước, phun gai mịn bằng súng chỉnh áp thấp. Phun 2 lớp mỏng thay vì một lớp dày.",
		Features:  []string{"Hệ nước 100% không mùi", "Gai mịn đều", "Chống trầy xước", "Dễ lau chùi"},
		Specs:     map[string]any{"coverage_m2_per_liter": 7.0, "recommended_coats": 2, "dry_time_touch_mins": 60, "dry_time_recoat_hours": 4, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "SP-02", "name": "Đen Gai Mịn", "hex": "#1F2937", "swatch_url": ""}},
		Coverage:  7.0, Featured: false,
	},
	{
		CategorySlug: "son-composite-nhua", SKU: "SL-CMP-001",
		Name: "Sơn Lót Composite Lotus Composite Primer", Slug: "son-lot-composite-lotus-composite-primer",
		Role: "primer", Finish: "matte", Usage: "both",
		Summary:   "Sơn lót cho FRP, nhựa ABS, PVC và chậu kiểng composite.",
		Desc:      "Lớp lót tăng bám dính trên bề mặt trơn, khó bám của composite và nhựa. Giúp lớp phủ không bong khi bề mặt co giãn nhẹ.",
		Guide:     "Chà nhám nhẹ, tẩy dầu mỡ, lau sạch. Sơn lót 1-2 lớp mỏng, để khô rồi sơn phủ.",
		Features:  []string{"Hệ nước 100% không mùi", "Bám dính trên nhựa", "Màng dẻo dai"},
		Specs:     map[string]any{"coverage_m2_per_liter": 8.0, "recommended_coats": 2, "dry_time_touch_mins": 45, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5-10%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{},
		Coverage:  8.0, Featured: false,
	},
	{
		CategorySlug: "son-composite-nhua", SKU: "SL-CMP-002",
		Name: "Sơn Phủ Composite Lotus Composite Topcoat", Slug: "son-phu-composite-lotus-composite-topcoat",
		Role: "topcoat", Finish: "satin", Usage: "both",
		Summary:   "Sơn phủ composite bán bóng, bền màu ngoài trời cho chậu kiểng và đồ FRP.",
		Desc:      "Sơn phủ cho chậu kiểng composite, đồ FRP và chi tiết nhựa ngoài trời. Màng dẻo, chống tia UV và chịu rửa trôi.",
		Guide:     "Sơn lót composite trước. Sơn phủ 2 lớp, tránh thi công khi bề mặt còn ẩm hoặc nắng gắt trực tiếp.",
		Features:  []string{"Hệ nước 100% không mùi", "Bám dính trên nhựa", "Kháng tia UV", "Chịu thời tiết"},
		Specs:     map[string]any{"coverage_m2_per_liter": 9.0, "recommended_coats": 2, "dry_time_touch_mins": 40, "dry_time_recoat_hours": 3, "dilution": "Nước sạch 5%"},
		Packaging: []string{"Lon 0.8L", "Lon 3.8L", "Thùng 18L / 20Kg"},
		Colors:    []map[string]string{{"code": "CP-01", "name": "Trắng Sứ", "hex": "#F8FAFC", "swatch_url": ""}, {"code": "CP-02", "name": "Xám Nhạt", "hex": "#E2E8F0", "swatch_url": ""}},
		Coverage:  9.0, Featured: true,
	},
}

type seedArticle struct {
	Type, Title, Slug, Excerpt, Content, RelatedSlug string
	Featured                                         bool
}

var seedArticles = []seedArticle{
	{
		Type: "project", Title: "Cổng và hàng rào sắt mạ kẽm giả gỗ cho nhà phố", Slug: "cong-sat-ma-kem-gia-go-nha-pho", RelatedSlug: "son-gia-go", Featured: true,
		Excerpt: "Hệ sơn giả gỗ biến cổng sắt mạ kẽm thành cổng vân gỗ ấm áp, bền ngoài trời mà không cần thay vật liệu.",
		Content: `Cổng và hàng rào sắt mạ kẽm có ưu điểm chắc chắn nhưng bề mặt kim loại thường lạnh và đơn điệu. Hệ sơn giả gỗ giải quyết đúng điểm này: giữ kết cấu sắt, khoác lên lớp vân gỗ mà mắt thường khó phân biệt với gỗ thật khi đứng ở khoảng cách thông thường.

Quy trình gồm bốn bước chính. Trước hết, bề mặt được vệ sinh, tẩy dầu mỡ và xử lý rỉ lỏng. Tiếp theo là lớp lót tạo nền đều màu, giúp vân gỗ lên rõ. Sau đó thợ vẽ vân gỗ bằng dụng cụ chuyên dụng, từng thớ theo chiều dài thanh sắt. Cuối cùng là lớp phủ bóng bảo vệ chống mưa nắng và tia UV.

Điểm cần lưu ý nhất là lớp nền và thời gian khô giữa các lớp. Nền không đều sẽ làm vân gỗ loang lổ, còn phủ bóng quá sớm có thể làm nhòe vân. Vì vậy mỗi lớp cần khô hoàn toàn trước khi sang bước tiếp theo, và nên thi công khi bề mặt khô ráo, tránh nắng gắt trực tiếp.

Khi bảo dưỡng, chỉ cần rửa nhẹ bằng nước sạch và lau khô. Nếu sau thời gian dài sử dụng lớp phủ ngoài trời bạc màu, có thể phủ lại lớp bóng bảo vệ mà không cần vẽ lại vân gỗ, miễn là lớp vân bên dưới còn nguyên vẹn.`,
	},
	{
		Type: "project", Title: "Tủ bếp gỗ tự nhiên sơn hệ nước: từ xử lý nền đến lớp phủ", Slug: "tu-bep-go-son-he-nuoc", RelatedSlug: "son-go-he-nuoc", Featured: false,
		Excerpt: "Tủ bếp là nơi tiếp xúc nước, dầu mỡ và nhiệt hằng ngày — hệ sơn gỗ hệ nước cần xử lý nền kỹ trước khi phủ.",
		Content: `Tủ bếp chịu tác động hằng ngày từ hơi nước, dầu mỡ và va chạm nhẹ. Vì vậy chất lượng của lớp sơn phụ thuộc nhiều vào khâu chuẩn bị hơn là vào số lớp phủ. Bề mặt càng sạch và phẳng thì màng sơn càng bám chắc và càng dễ lau chùi về sau.

Bắt đầu bằng việc tháo tay nắm và vệ sinh toàn bộ cánh tủ bằng khăn ẩm, sau đó chà nhám nhẹ theo thớ gỗ để mở bề mặt. Những vết ố dầu mỡ cần được tẩy sạch trước khi sơn lót, vì dầu mỡ còn sót sẽ làm sơn lót bong từng mảng.

Lớp lót có nhiệm vụ lấp thớ gỗ và làm đều màu nền. Sau khi lót, chà nhám nhẹ một lần nữa để loại bỏ xơ gỗ dựng lên, lau bụi thật kỹ rồi mới sơn phủ. Sơn phủ nên thi công thành hai lớp mỏng thay vì một lớp dày, mỗi lớp cách nhau đủ thời gian khô theo hướng dẫn của sản phẩm.

Sau khi hoàn thiện, nên chờ màng sơn cứng hẳn trước khi lắp tay nắm và sử dụng mạnh. Việc lau chùi hằng ngày chỉ cần khăn mềm ẩm, tránh chất tẩy mạnh lên bề mặt mới sơn trong thời gian đầu.`,
	},
	{
		Type: "project", Title: "Làm mới mái ngói và sân bê tông sau mùa mưa", Slug: "lam-moi-mai-ngoi-san-be-tong", RelatedSlug: "son-ngoi-be-tong", Featured: false,
		Excerpt: "Sau mùa mưa, mái ngói và sân bê tông thường bám rêu mốc và bạc màu — xử lý đúng thứ tự giúp lớp sơn mới bám bền.",
		Content: `Mùa mưa để lại rêu mốc, bụi bẩn và vệt nước trên mái ngói và sân bê tông. Nếu sơn phủ trực tiếp lên lớp bẩn này, sơn sẽ bám vào rêu mốc chứ không bám vào bề mặt, và sẽ bong rất nhanh khi gặp mưa trở lại.

Bước đầu tiên luôn là làm sạch cơ học: rửa áp lực nhẹ hoặc chà bằng bàn chải cứng, loại bỏ rêu mốc và mảng bám lỏng. Sau khi rửa, bề mặt cần được để khô hoàn toàn — sơn lên bề mặt còn ẩm là nguyên nhân phổ biến gây phồng rộp.

Tiếp theo là lớp lót kháng kiềm cho bê tông và ngói mới hoặc đã bạc màu. Lớp lót giúp lớp phủ không bị kiềm trong xi măng làm phấn hóa sớm. Cuối cùng mới đến hai lớp sơn phủ, thi công từ trên xuống dưới đối với mái, và theo từng khoang đối với sân để tránh giẫm lên lớp sơn còn ướt.

Về lâu dài, nên kiểm tra định kỳ các vị trí đọng nước và rãnh thoát nước. Sơn tốt cũng khó bền nếu nước đọng lâu ngày trên bề mặt, vì vậy xử lý thoát nước luôn đi cùng với sơn phủ.`,
	},
	{
		Type: "tech_support", Title: "Chuẩn bị bề mặt gỗ trước khi sơn hệ nước", Slug: "chuan-bi-be-mat-go-truoc-khi-son", RelatedSlug: "son-go-he-nuoc", Featured: true,
		Excerpt: "Chà nhám, làm sạch và xử lý ẩm đúng cách quyết định phần lớn độ bám và độ mịn của lớp sơn gỗ hệ nước.",
		Content: `Bề mặt gỗ trước khi sơn thường có ba vấn đề: bụi gỗ, xơ gỗ dựng lên và vết bẩn. Chà nhám là cách xử lý cả ba, nhưng cần đúng hạt nhám và đúng chiều. Chà theo thớ gỗ, bắt đầu từ hạt thô cho bề mặt xấu rồi chuyển dần sang hạt mịn hơn, tránh chà ngang thớ vì sẽ để lại vết xước khó che.

Sau khi chà, lau sạch bụi bằng khăn ẩm vắt khô hoặc khí nén nhẹ. Bụi còn sót sẽ nổi lên trong lớp sơn lót thành hạt sần. Nếu gỗ từng dính dầu mỡ hoặc keo, cần tẩy sạch điểm đó trước, vì sơn hệ nước bám kém trên bề mặt còn dầu.

Độ ẩm của gỗ cũng cần được kiểm tra bằng cảm quan và thời gian phơi khô: gỗ còn ẩm sẽ làm sơn chậm khô và dễ phồng. Nên thi công khi gỗ đã khô ráo, nhiệt độ và độ ẩm không khí ở mức bình thường, tránh sơn khi trời nồm ẩm.

Cuối cùng, luôn sơn thử trên một mảnh gỗ cùng loại trước khi sơn toàn bộ. Mảnh thử giúp kiểm tra màu sắc, độ bóng và độ bám, tránh phải tẩy toàn bộ khi màu lên không như mong muốn.`,
	},
	{
		Type: "tech_support", Title: "Vẽ vân gỗ trên sắt mạ kẽm: các bước cơ bản", Slug: "ve-van-go-tren-sat-ma-kem", RelatedSlug: "son-gia-go", Featured: true,
		Excerpt: "Vân gỗ đẹp đến từ lớp nền đều, dụng cụ vẽ phù hợp và nhịp tay ổn định — không phải từ một lớp sơn duy nhất.",
		Content: `Vẽ vân gỗ là kỹ thuật tạo hiệu ứng, không phải một loại sơn đơn lẻ. Toàn bộ hệ gồm lớp lót nền, lớp tạo vân và lớp phủ bảo vệ, mỗi lớp có nhiệm vụ riêng. Bỏ qua lớp lót hoặc phủ bóng quá sớm đều làm hỏng kết quả cuối cùng.

Bắt đầu với bề mặt sắt mạ kẽm đã vệ sinh và xử lý rỉ lỏng. Lớp lót nền được sơn đều tay để tạo màu gỗ nền — thường là tông sáng hơn màu vân. Khi lớp nền khô hoàn toàn, dùng dụng cụ vẽ vân kéo theo chiều dài của thanh sắt, tạo các thớ dài, hơi cong tự nhiên, tránh vẽ thành các đoạn ngắn đứt quãng.

Sau khi vẽ vân, cần chờ vân khô hẳn trước khi phủ bóng. Lớp phủ bóng có hai nhiệm vụ: bảo vệ vân khỏi mưa nắng và tạo độ sâu cho bề mặt. Phủ thành hai lớp mỏng, lớp sau cách lớp trước đủ thời gian khô.

Lỗi thường gặp là vẽ vân khi nền còn ướt, làm vân nhòe, hoặc vẽ quá dày làm vân trông giả. Cách khắc phục là sơn thử trên một thanh sắt mẫu, điều chỉnh lực tay và lượng sơn trên dụng cụ trước khi làm trên công trình thật.`,
	},
	{
		Type: "tech_support", Title: "Xử lý rỉ sét trên kim loại trước khi sơn phủ", Slug: "xu-ly-ri-set-tren-kim-loai", RelatedSlug: "son-kim-loai-he-nuoc", Featured: false,
		Excerpt: "Rỉ sét phải được loại bỏ phần lỏng và xử lý phần bám chặt — sơn phủ lên rỉ lỏng chỉ che tạm thời.",
		Content: `Rỉ sét gồm hai phần: lớp rỉ lỏng, bong tróc và lớp rỉ bám chặt vào kim loại. Sơn phủ chỉ bám được khi lớp rỉ lỏng đã được loại bỏ. Vì vậy bước đầu tiên luôn là làm sạch cơ học bằng bàn chải sắt, giấy nhám hoặc máy chà, cho đến khi bề mặt không còn mảng rỉ bong ra khi chạm tay.

Sau khi chà, tẩy dầu mỡ bằng dung dịch tẩy rửa phù hợp và lau khô. Dầu mỡ là nguyên nhân ít được chú ý nhưng gây bong sơn rất phổ biến, đặc biệt trên khung sắt mới gia công hoặc vừa hàn.

Tiếp theo là lớp lót chống rỉ. Lớp lót này bám vào kim loại sạch và ức chế quá trình rỉ tiếp diễn dưới màng sơn. Chỉ khi lớp lót đã khô mới sơn phủ màu. Sơn phủ 2 trong 1 có thể rút ngắn quy trình trên bề mặt đã làm sạch tốt, nhưng không thay thế được bước làm sạch rỉ lỏng.

Về lâu dài, các vị trí hàn, mối nối và chỗ đọng nước cần được kiểm tra trước tiên, vì đó là nơi rỉ quay lại sớm nhất. Xử lý sớm một điểm rỉ nhỏ luôn rẻ và nhanh hơn sơn lại toàn bộ bề mặt.`,
	},
	{
		Type: "tech_support", Title: "Sơn mái ngói: vệ sinh, lót kháng kiềm và phủ", Slug: "son-mai-ngoi-ve-sinh-lot-khang-kiem", RelatedSlug: "son-ngoi-be-tong", Featured: false,
		Excerpt: "Sơn mái ngói bền hay không phụ thuộc vào vệ sinh, lớp lót kháng kiềm và thời tiết lúc thi công.",
		Content: `Mái ngói tích tụ rêu mốc, bụi và lớp phấn hóa theo thời gian. Vệ sinh là bước không thể bỏ qua: rửa và chà sạch rêu mốc, thay những viên ngói nứt vỡ trước khi sơn. Sơn lên ngói nứt chỉ che tạm thời và nước sẽ tiếp tục thấm qua vết nứt.

Sau vệ sinh, bề mặt phải khô hoàn toàn. Với ngói và bê tông, lớp lót kháng kiềm đóng vai trò quan trọng vì kiềm trong xi măng có thể làm lớp phủ phấn hóa và bạc màu sớm. Một lớp lót kháng kiềm mỏng, đều sẽ kéo dài tuổi thọ của lớp phủ đáng kể.

Sơn phủ nên thi công thành hai lớp, từ đỉnh mái xuống, tránh đi lại trên lớp sơn còn ướt. Thời tiết lý tưởng là ngày khô ráo, không mưa trong ít nhất một ngày sau khi phủ lớp cuối. Thi công khi trời sắp mưa hoặc bề mặt còn ẩm là nguyên nhân hàng đầu gây phồng rộp và bong tróc.

Cuối cùng, cần kiểm tra hệ thống thoát nước và máng xối sau khi sơn. Sơn phủ bảo vệ bề mặt, nhưng không thay thế được mái dột do ngói vỡ hoặc máng tắc.`,
	},
	{
		Type: "solution", Title: "Giải pháp sơn gỗ cho đồ nội thất xuất khẩu", Slug: "giai-phap-son-go-noi-that-xuat-khau", RelatedSlug: "son-go-he-nuoc", Featured: true,
		Excerpt: "Đồ nội thất xuất khẩu cần hệ sơn ổn định theo lô, màu đồng đều và quy trình ghi chép rõ ràng từ lót đến phủ.",
		Content: `Đồ nội thất xuất khẩu đặt ra yêu cầu cao hơn hàng nội địa ở ba điểm: màu đồng đều giữa các lô, màng sơn bền trong vận chuyển, và hồ sơ quy trình rõ ràng để đối tác kiểm tra. Hệ sơn gỗ hệ nước đáp ứng tốt các yêu cầu này khi quy trình được chuẩn hóa, vì sơn hệ nước ít mùi và khô nhanh hơn, giúp rút ngắn chu kỳ sản xuất.

Chuẩn hóa bắt đầu từ bề mặt: cùng một quy trình chà nhám, cùng số lớp lót và cùng thời gian khô cho mọi lô. Màu sắc cần được đối chiếu trên mẫu chuẩn trước mỗi lô, thay vì dựa vào cảm quan của từng thợ. Việc này nghe đơn giản nhưng quyết định việc màu của lô sau có khớp lô trước hay không.

Về màng sơn, đồ nội thất xuất khẩu chịu rung lắc và thay đổi nhiệt độ trong vận chuyển. Lớp phủ cần đủ cứng để chống trầy xước nhẹ, nhưng vẫn dẻo để không nứt khi gỗ co giãn theo độ ẩm. Thi công hai lớp mỏng thường cho kết quả ổn định hơn một lớp dày.

Cuối cùng, trước khi xuất hàng, cần kiểm tra ngẫu nhiên một tỷ lệ sản phẩm theo lô: độ bám, độ bóng và màu sắc. Những lô không đạt cần được xử lý trước khi đóng gói, vì chi phí sửa chữa sau khi hàng đã lên container luôn cao hơn nhiều.`,
	},
	{
		Type: "solution", Title: "Sơn không mùi cho không gian đang sử dụng", Slug: "son-khong-mui-cho-khong-gian-dang-su-dung", RelatedSlug: "", Featured: false,
		Excerpt: "Sơn hệ nước ít mùi giúp thi công trong không gian đang sử dụng dễ chịu hơn, nhưng vẫn cần thông gió và phân khu thi công.",
		Content: `Khi cần sơn lại văn phòng, lớp học hay nhà ở mà không thể dọn trống hoàn toàn, yếu tố quan trọng nhất là mùi sơn và thời gian khô. Sơn hệ nước ít mùi hơn sơn dung môi, vì vậy người sử dụng không gian cảm thấy dễ chịu hơn trong và sau khi thi công. Tuy nhiên ít mùi không có nghĩa là không cần thông gió.

Cách làm thực tế là thi công theo từng khu vực, luân phiên đóng và mở từng phòng. Khu vực đang sơn cần được thông gió tốt, còn khu vực đang sử dụng nên được tách bằng cửa và rèm che bụi. Lịch thi công nên sắp xếp vào cuối ngày làm việc hoặc cuối tuần cho các khu vực nhạy cảm.

Chọn hệ sơn cũng cần cân nhắc bề mặt: gỗ, kim loại và tường mỗi loại cần hệ lót và phủ khác nhau. Dùng đúng hệ cho từng bề mặt giúp rút ngắn số lớp và thời gian thi công, từ đó rút ngắn thời gian phải phân khu.

Cuối cùng, nên thông báo trước lịch sơn cho người sử dụng không gian, kèm thời gian dự kiến có thể dùng lại từng khu vực. Việc này giúp mọi người chủ động sắp xếp và tránh chạm vào lớp sơn còn ướt.`,
	},
	{
		Type: "solution", Title: "Giải pháp sơn thùng loa sân khấu và hội trường", Slug: "giai-phap-son-thung-loa-san-khau", RelatedSlug: "son-loa", Featured: false,
		Excerpt: "Thùng loa sân khấu cần bề mặt gai sần cứng, chống trầy xước và chịu vận chuyển — quy trình lót và phun gai quyết định chất lượng.",
		Content: `Thùng loa sân khấu và hội trường bị va đập thường xuyên khi vận chuyển, lắp đặt và tháo dỡ. Vì vậy bề mặt sơn loa không chỉ cần đẹp mà còn phải cứng và bám chắc. Hệ sơn tạo gai sần đáp ứng yêu cầu này: lớp gai nổi giúp che vết xước nhỏ và phân tán lực va chạm tốt hơn bề mặt phẳng.

Quy trình gồm lót, tạo gai và phủ bảo vệ. Lớp lót làm đều bề mặt gỗ hoặc MDF, giảm hút sơn và giúp gai bám đều. Lớp gai được phun bằng súng chuyên dụng, điều chỉnh áp lực để hạt gai đều tay — áp quá cao làm gai thưa, áp quá thấp làm gai vón cục.

Với thùng loa dùng trong nhà, gai mịn mang lại cảm giác tinh tế hơn, trong khi sân khấu ngoài trời thường ưu tiên gai thô hơn để che va đập. Việc chọn độ gai nên dựa vào môi trường sử dụng thực tế, không nên chọn theo cảm quan trên một mẫu nhỏ duy nhất.

Sau khi hoàn thiện, thùng loa cần được để khô và cứng hẳn trước khi xếp chồng vận chuyển. Xếp chồng quá sớm là nguyên nhân phổ biến gây dính và tróc lớp gai ở các điểm tiếp xúc.`,
	},
	{
		Type: "news", Title: "Sơn Lotus mở rộng hệ sản phẩm composite và chống thấm", Slug: "son-lotus-mo-rong-he-composite-chong-tham", RelatedSlug: "son-composite-nhua", Featured: false,
		Excerpt: "Danh mục Sơn Lotus nay có đầy đủ hệ lót và phủ cho composite, nhựa, cùng các dòng trám trét và chống thấm hệ nước.",
		Content: `Danh mục sản phẩm Sơn Lotus vừa được bổ sung hệ lót và phủ dành cho composite và nhựa, gồm sơn lót composite và sơn phủ composite bán bóng. Hai dòng này được phát triển cho chậu kiểng composite, đồ FRP và chi tiết nhựa ngoài trời, nơi bề mặt trơn và khó bám là thách thức chính.

Cùng với hệ composite, nhóm trám trét và chống thấm tiếp tục được hoàn thiện với bột trét gỗ hệ nước và keo bóng nước chống thấm. Nhờ vậy, một công trình có thể dùng chung một hệ sơn nước từ khâu trám trét, lót, phủ màu đến phủ bảo vệ, thay vì phải phối hợp nhiều hệ sơn khác nhau.

Khách hàng có thể xem toàn bộ hệ sản phẩm mới trong danh mục trên website, kèm thông số độ phủ, quy cách đóng gói và bảng màu của từng dòng. Đội ngũ kỹ thuật cũng sẵn sàng tư vấn chọn hệ lót và phủ phù hợp với từng loại bề mặt cụ thể.

Thông tin chi tiết về từng sản phẩm, hướng dẫn thi công và yêu cầu báo giá đều có trên trang sản phẩm tương ứng hoặc qua form liên hệ trên website.`,
	},
	{
		Type: "news", Title: "Nhận mẫu sơn miễn phí và gửi yêu cầu báo giá như thế nào", Slug: "nhan-mau-son-mien-phi-va-bao-gia", RelatedSlug: "", Featured: false,
		Excerpt: "Ba bước để nhận mẫu sơn và báo giá: chọn sản phẩm, gửi yêu cầu qua form hoặc hotline, và chờ kỹ thuật tư vấn.",
		Content: `Để nhận mẫu sơn, khách hàng chỉ cần xác định bề mặt cần sơn và dòng sản phẩm quan tâm, sau đó gửi yêu cầu qua form liên hệ trên website, qua Zalo hoặc gọi hotline 0943 966 662. Trong yêu cầu, nên ghi rõ loại bề mặt, diện tích ước tính và màu sắc mong muốn để kỹ thuật tư vấn chính xác hơn.

Sau khi nhận yêu cầu, đội ngũ kỹ thuật sẽ liên hệ lại trong giờ làm việc để xác nhận màu sắc, quy cách đóng gói và địa chỉ nhận mẫu. Việc ghi rõ diện tích cũng giúp tính toán lượng sơn cần dùng, tránh mua thừa hoặc thiếu khi thi công.

Đối với yêu cầu báo giá số lượng lớn hoặc cho công trình, khách hàng nên gửi kèm thông tin về bề mặt, số lượng và tiến độ dự kiến. Những thông tin này giúp báo giá sát thực tế hơn và rút ngắn thời gian trao đổi qua lại.

Mọi yêu cầu gửi qua form, email sales@sonlotus.vn hoặc Zalo đều được tiếp nhận chung một đầu mối, vì vậy khách hàng chỉ cần chọn kênh thuận tiện nhất mà không cần gửi trùng lặp nhiều nơi.`,
	},
}

// Seed is idempotent backfill: for every seed category, product and
// article, insert only if its slug does not already exist. It never
// overwrites admin-edited rows. One shared function for server boot and
// migrator --seed-only (PLAN §3). After seeding it runs the FTS backfill.
// The old behaviour (no-op when categories exist) would hide new seed
// data from databases seeded before the new items existed.
func Seed(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	catIDs := map[string]int64{}
	for _, c := range seedCategories {
		var id int64
		err := tx.QueryRow(`SELECT id FROM categories WHERE slug = ?`, c.Slug).Scan(&id)
		if err == nil {
			catIDs[c.Slug] = id
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		res, err := tx.Exec(`INSERT INTO categories(name, slug, short_description, description_html, surface_type, icon_name, sort_order, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, 1)`, c.Name, c.Slug, c.Short, c.Desc, c.Surface, c.Icon, c.Sort)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.Slug, err)
		}
		id, err = res.LastInsertId()
		if err != nil {
			return err
		}
		catIDs[c.Slug] = id
	}
	allProducts := append(append([]seedProduct{}, seedProducts...), seedProductsExtra...)
	for _, pr := range allProducts {
		var n int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM products WHERE slug = ?`, pr.Slug).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		catID, ok := catIDs[pr.CategorySlug]
		if !ok {
			return fmt.Errorf("seed product %s: unknown category %s", pr.Slug, pr.CategorySlug)
		}
		colors := pr.Colors
		if colors == nil {
			colors = []map[string]string{}
		}
		if _, err := tx.Exec(`INSERT INTO products(category_id, sku, name, slug, coating_role, finish_type, usage_area,
			short_summary, description_html, application_guide_html, features_json, specs_json, packaging_json, colors_json, images_json,
			coverage_rate, is_featured, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
			catID, pr.SKU, pr.Name, pr.Slug, pr.Role, pr.Finish, pr.Usage,
			pr.Summary, pr.Desc, pr.Guide, mustJSON(pr.Features), mustJSON(pr.Specs), mustJSON(pr.Packaging), mustJSON(colors), mustJSON([]string{}),
			pr.Coverage, boolToInt(pr.Featured)); err != nil {
			return fmt.Errorf("seed product %s: %w", pr.Slug, err)
		}
	}
	for _, a := range seedArticles {
		var n int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM articles WHERE slug = ?`, a.Slug).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		var relID any
		if a.RelatedSlug != "" {
			id, ok := catIDs[a.RelatedSlug]
			if !ok {
				return fmt.Errorf("seed article %s: unknown category %s", a.Slug, a.RelatedSlug)
			}
			relID = id
		}
		if _, err := tx.Exec(`INSERT INTO articles(type, title, slug, excerpt, content_html, thumbnail_url, author_name, related_category_id, is_featured, is_published, seo_title, seo_description)
			VALUES (?, ?, ?, ?, ?, '', 'Kỹ Thuật Sơn Lotus', ?, ?, 1, '', '')`,
			a.Type, a.Title, a.Slug, a.Excerpt, a.Content, relID, boolToInt(a.Featured)); err != nil {
			return fmt.Errorf("seed article %s: %w", a.Slug, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return BackfillFTS(db)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
