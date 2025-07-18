// Code generated by codegen. DO NOT EDIT.

package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"unified-ads-mcp/internal/facebook/generated/common"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// user_create_businesseArgs defines the typed arguments for user_create_businesse
type user_create_businesseArgs struct {
	ID                      string `json:"id" jsonschema:"required,description=User ID,pattern=^[0-9]+$"`
	ChildBusinessExternalId string `json:"child_business_external_id,omitempty" jsonschema:"description=ID of the Child Business External,pattern=^[0-9]+$"`
	Email                   string `json:"email,omitempty" jsonschema:"description=Email"`
	Name                    string `json:"name" jsonschema:"description=Name,required"`
	PrimaryPage             string `json:"primary_page,omitempty" jsonschema:"description=Primary Page"`
	SalesRepEmail           string `json:"sales_rep_email,omitempty" jsonschema:"description=Sales Rep Email"`
	SurveyBusinessType      string `json:"survey_business_type,omitempty" jsonschema:"description=Survey Business Type"`
	SurveyNumAssets         int    `json:"survey_num_assets,omitempty" jsonschema:"description=Survey Num Assets"`
	SurveyNumPeople         int    `json:"survey_num_people,omitempty" jsonschema:"description=Survey Num People"`
	TimezoneId              string `json:"timezone_id,omitempty" jsonschema:"description=ID of the Timezone,pattern=^[0-9]+$"`
	Vertical                string `json:"vertical" jsonschema:"description=Vertical,required"`
}

// RegisterUserCreateBusinesseHandler registers the user_create_businesse tool
func RegisterUserCreateBusinesseHandler(s *server.MCPServer) error {
	tool := mcp.NewToolWithRawSchema(
		"user_create_businesse",
		"Create or update businesses for this User Returns Business. Required: name, vertical (enum)",
		json.RawMessage(`{"additionalProperties":false,"properties":{"child_business_external_id":{"description":"ID of the Child Business External","pattern":"^[0-9]+$","type":"string"},"email":{"description":"Email","type":"string"},"id":{"description":"User ID","pattern":"^[0-9]+$","type":"string"},"name":{"description":"Name","type":"string"},"primary_page":{"description":"Primary Page","type":"string"},"sales_rep_email":{"description":"Sales Rep Email","type":"string"},"survey_business_type":{"description":"Survey Business Type (enum: userbusinesses_survey_business_type_enum_param)","enum":["ADVERTISER","AGENCY","APP_DEVELOPER","PUBLISHER"],"type":"string"},"survey_num_assets":{"description":"Survey Num Assets","type":"integer"},"survey_num_people":{"description":"Survey Num People","type":"integer"},"timezone_id":{"description":"ID of the Timezone (enum: userbusinesses_timezone_id_enum_param)","enum":["0","1","2","3","4","5","6","7","8","9","10","11","12","13","14","15","16","17","18","19","20","21","22","23","24","25","26","27","28","29","30","31","32","33","34","35","36","37","38","39","40","41","42","43","44","45","46","47","48","49","50","51","52","53","54","55","56","57","58","59","60","61","62","63","64","65","66","67","68","69","70","71","72","73","74","75","76","77","78","79","80","81","82","83","84","85","86","87","88","89","90","91","92","93","94","95","96","97","98","99","100","101","102","103","104","105","106","107","108","109","110","111","112","113","114","115","116","117","118","119","120","121","122","123","124","125","126","127","128","129","130","131","132","133","134","135","136","137","138","139","140","141","142","143","144","145","146","147","148","149","150","151","152","153","154","155","156","157","158","159","160","161","162","163","164","165","166","167","168","169","170","171","172","173","174","175","176","177","178","179","180","181","182","183","184","185","186","187","188","189","190","191","192","193","194","195","196","197","198","199","200","201","202","203","204","205","206","207","208","209","210","211","212","213","214","215","216","217","218","219","220","221","222","223","224","225","226","227","228","229","230","231","232","233","234","235","236","237","238","239","240","241","242","243","244","245","246","247","248","249","250","251","252","253","254","255","256","257","258","259","260","261","262","263","264","265","266","267","268","269","270","271","272","273","274","275","276","277","278","279","280","281","282","283","284","285","286","287","288","289","290","291","292","293","294","295","296","297","298","299","300","301","302","303","304","305","306","307","308","309","310","311","312","313","314","315","316","317","318","319","320","321","322","323","324","325","326","327","328","329","330","331","332","333","334","335","336","337","338","339","340","341","342","343","344","345","346","347","348","349","350","351","352","353","354","355","356","357","358","359","360","361","362","363","364","365","366","367","368","369","370","371","372","373","374","375","376","377","378","379","380","381","382","383","384","385","386","387","388","389","390","391","392","393","394","395","396","397","398","399","400","401","402","403","404","405","406","407","408","409","410","411","412","413","414","415","416","417","418","419","420","421","422","423","424","425","426","427","428","429","430","431","432","433","434","435","436","437","438","439","440","441","442","443","444","445","446","447","448","449","450","451","452","453","454","455","456","457","458","459","460","461","462","463","464","465","466","467","468","469","470","471","472","473","474","475","476","477","478","479","480"],"type":"string"},"vertical":{"description":"Vertical (enum: userbusinesses_vertical_enum_param)","enum":["ADVERTISING","AUTOMOTIVE","CONSUMER_PACKAGED_GOODS","ECOMMERCE","EDUCATION","ENERGY_AND_UTILITIES","ENTERTAINMENT_AND_MEDIA","FINANCIAL_SERVICES","GAMING","GOVERNMENT_AND_POLITICS","HEALTH","LUXURY","MARKETING","NON_PROFIT","NOT_SET","ORGANIZATIONS_AND_ASSOCIATIONS","OTHER","PROFESSIONAL_SERVICES","RESTAURANT","RETAIL","TECHNOLOGY","TELECOM","TRAVEL"],"type":"string"}},"required":["id","name","vertical"],"type":"object"}`),
	)

	s.AddTool(tool, UserCreateBusinesseHandler)
	return nil
}

// UserCreateBusinesseHandler handles the user_create_businesse tool
func UserCreateBusinesseHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args user_create_businesseArgs
	if err := request.BindArguments(&args); err != nil {
		return common.HandleBindError(err)
	}
	endpoint := fmt.Sprintf("/%s/businesses", args.ID)
	// Prepare request body
	body := make(map[string]interface{})
	if args.ChildBusinessExternalId != "" {
		body["child_business_external_id"] = args.ChildBusinessExternalId
	}
	if args.Email != "" {
		body["email"] = args.Email
	}
	if args.Name != "" {
		body["name"] = args.Name
	}
	if args.PrimaryPage != "" {
		body["primary_page"] = args.PrimaryPage
	}
	if args.SalesRepEmail != "" {
		body["sales_rep_email"] = args.SalesRepEmail
	}
	if args.SurveyBusinessType != "" {
		body["survey_business_type"] = args.SurveyBusinessType
	}
	if args.SurveyNumAssets > 0 {
		body["survey_num_assets"] = args.SurveyNumAssets
	}
	if args.SurveyNumPeople > 0 {
		body["survey_num_people"] = args.SurveyNumPeople
	}
	if args.TimezoneId != "" {
		body["timezone_id"] = args.TimezoneId
	}
	if args.Vertical != "" {
		body["vertical"] = args.Vertical
	}

	result, err := common.MakeGraphAPIRequest(ctx, "POST", endpoint, nil, body)

	if err != nil {
		return common.HandleAPIError(err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(string(result)),
		},
	}, nil
}
