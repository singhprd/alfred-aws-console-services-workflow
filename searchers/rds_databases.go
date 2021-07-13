package searchers

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	aw "github.com/deanishe/awgo"
	"github.com/rkoval/alfred-aws-console-services-workflow/awsworkflow"
	"github.com/rkoval/alfred-aws-console-services-workflow/caching"
	"github.com/rkoval/alfred-aws-console-services-workflow/util"
)

type RDSDatabaseSearcher struct{}

func (s RDSDatabaseSearcher) Search(wf *aw.Workflow, query string, cfg aws.Config, forceFetch bool, fullQuery string) error {
	cacheName := util.GetCurrentFilename()
	es := caching.LoadRdsDBInstanceArrayFromCache(wf, cfg, cacheName, s.fetch, forceFetch, fullQuery)
	for _, entity := range es {
		s.addToWorkflow(wf, query, cfg, entity)
	}
	return nil
}

func (s RDSDatabaseSearcher) fetch(cfg aws.Config) ([]types.DBInstance, error) {
	svc := rds.NewFromConfig(cfg)

	pageToken := ""
	var entities []types.DBInstance
	for {
		params := &rds.DescribeDBInstancesInput{
			MaxRecords: aws.Int32(100),
		}
		if pageToken != "" {
			params.Marker = aws.String(pageToken)
		}
		resp, err := svc.DescribeDBInstances(context.TODO(), params)
		if err != nil {
			return nil, err
		}

		entities = append(entities, resp.DBInstances...)

		if resp.Marker != nil {
			pageToken = *resp.Marker
		} else {
			break
		}
	}

	return entities, nil
}

func (s RDSDatabaseSearcher) addToWorkflow(wf *aw.Workflow, query string, config aws.Config, entity types.DBInstance) {
	subtitleArray := []string{}
	var engineString string
	if entity.Engine != nil && *entity.Engine != "" {
		engineString += *entity.Engine
	}
	if entity.EngineVersion != nil && *entity.EngineVersion != "" {
		engineString += " " + *entity.EngineVersion
	}
	subtitleArray = util.AppendString(subtitleArray, &engineString)
	subtitleArray = util.AppendString(subtitleArray, entity.DBInstanceClass)

	title := *entity.DBInstanceIdentifier
	if entity.DBName != nil && *entity.DBName != title {
		subtitleArray = util.AppendString(subtitleArray, entity.DBName)
	}

	subtitle := strings.Join(subtitleArray, " – ")

	util.NewURLItem(wf, title).
		Subtitle(subtitle).
		Arg(fmt.Sprintf(
			"https://%s.console.aws.amazon.com/rds/home?region=%s#database:id=%s;is-cluster=false",
			config.Region,
			config.Region,
			*entity.DBInstanceIdentifier,
		)).
		Icon(awsworkflow.GetImageIcon("rds")).
		Valid(true)
}