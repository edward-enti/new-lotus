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

// Seed inserts the SPEC §5 categories and products if and only if the
// categories table is empty, in ONE transaction. Otherwise it is a no-op
// (idempotent). One shared function for server boot and migrator
// --seed-only (PLAN §3). After seeding it runs the FTS backfill.
func Seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	catIDs := map[string]int64{}
	for _, c := range seedCategories {
		res, err := tx.Exec(`INSERT INTO categories(name, slug, short_description, description_html, surface_type, icon_name, sort_order, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, 1)`, c.Name, c.Slug, c.Short, c.Desc, c.Surface, c.Icon, c.Sort)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", c.Slug, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		catIDs[c.Slug] = id
	}
	for _, p := range seedProducts {
		catID, ok := catIDs[p.CategorySlug]
		if !ok {
			return fmt.Errorf("seed product %s: unknown category %s", p.Slug, p.CategorySlug)
		}
		images := []string{}
		if p.Colors == nil {
			p.Colors = []map[string]string{}
		}
		_, err := tx.Exec(`INSERT INTO products(category_id, sku, name, slug, coating_role, finish_type, usage_area,
			short_summary, description_html, application_guide_html, features_json, specs_json, packaging_json, colors_json, images_json,
			coverage_rate, is_featured, is_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
			catID, p.SKU, p.Name, p.Slug, p.Role, p.Finish, p.Usage,
			p.Summary, p.Desc, p.Guide, mustJSON(p.Features), mustJSON(p.Specs), mustJSON(p.Packaging), mustJSON(p.Colors), mustJSON(images),
			p.Coverage, boolToInt(p.Featured))
		if err != nil {
			return fmt.Errorf("seed product %s: %w", p.Slug, err)
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
