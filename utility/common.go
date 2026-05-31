package utility

import "github.com/milvus-io/milvus-sdk-go/v2/entity"

const (
	MilvusDBName         = "agent"
	MilvusCollectionName = "paper"
)

var FileDir = "./papers/"

var Fields = []*entity.Field{
	// id (主键)
	{
		Name:       "id",
		DataType:   entity.FieldTypeVarChar,
		PrimaryKey: true,
		TypeParams: map[string]string{"max_length": "256"},
	},
	// paper_id (VarChar)
	{
		Name:       "paper_id",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "128"},
	},
	// title (VarChar)
	{
		Name:       "title",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "1024"},
	},
	// authors (VarChar，存储JSON字符串)
	{
		Name:       "authors",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "2048"},
	},
	// abstract (VarChar)
	{
		Name:       "abstract",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "65535"},
	},
	// title_vector (FloatVector, dim=1024)
	{
		Name:       "title_vector",
		DataType:   entity.FieldTypeFloatVector,
		TypeParams: map[string]string{"dim": "1024"},
	},
	// abstract_vector (FloatVector)
	{
		Name:       "abstract_vector",
		DataType:   entity.FieldTypeFloatVector,
		TypeParams: map[string]string{"dim": "1024"},
	},
	// topic_vector (FloatVector)
	{
		Name:       "topic_vector",
		DataType:   entity.FieldTypeFloatVector,
		TypeParams: map[string]string{"dim": "1024"},
	},
	// keywords (Array<VarChar>)
	{
		Name:     "keywords",
		DataType: entity.FieldTypeArray,
		TypeParams: map[string]string{
			"max_capacity": "50",
		},
		ElementType: entity.FieldTypeVarChar,
	},
	// publication_year (Int32)
	{
		Name:     "publication_year",
		DataType: entity.FieldTypeInt32,
	},
	// citation_count (Int32)
	{
		Name:     "citation_count",
		DataType: entity.FieldTypeInt32,
	},
	// venue (VarChar)
	{
		Name:       "venue",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "256"},
	},
	// full_text (VarChar，存储全文或分块)
	{
		Name:       "full_text",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "65535"},
	},
	// full_text_vector (FloatVector)
	{
		Name:       "full_text_vector",
		DataType:   entity.FieldTypeFloatVector,
		TypeParams: map[string]string{"dim": "1024"},
	},
	// references (VarChar，JSON字符串)
	{
		Name:       "references",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "65535"},
	},
	// methodology (VarChar)
	{
		Name:       "methodology",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "8192"},
	},
	// dataset_info (VarChar)
	{
		Name:       "dataset_info",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "4096"},
	},
	// innovation_points (VarChar)
	{
		Name:       "innovation_points",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "8192"},
	},
	// field (VarChar)
	{
		Name:       "field",
		DataType:   entity.FieldTypeVarChar,
		TypeParams: map[string]string{"max_length": "256"},
	},
	// metadata (JSON)
	{
		Name:     "metadata",
		DataType: entity.FieldTypeJSON,
	},
}
