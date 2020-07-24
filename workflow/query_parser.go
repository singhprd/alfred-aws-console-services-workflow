package workflow

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	aw "github.com/deanishe/awgo"
	"github.com/rkoval/alfred-aws-console-services-workflow/core"
)

func ParseQueryAndPopulateItems(wf *aw.Workflow, awsServices []core.AwsService, query string, session *session.Session, forceFetch bool) string {
	// TODO break apart this function
	// TODO add better lexing here to route populators

	fullQuery := query

	splitQuery := strings.Split(query, " ")
	if len(splitQuery) > 1 {
		id := splitQuery[0]
		var awsService *core.AwsService
		for i := range awsServices {
			if awsServices[i].Id == id {
				awsService = &awsServices[i]
				break
			}
		}

		if awsService != nil {
			query = strings.Join(splitQuery[1:], " ")
			populater := PopulatersByServiceId[id]
			if strings.HasPrefix(query, "$") && populater != nil {
				query = query[1:]
				log.Printf("using populater associated with %s", id)
				err := populater(wf, query, session, forceFetch, fullQuery)
				if err != nil {
					wf.FatalError(err)
				}
				return query
			} else {
				// prepend the home to the sub-service list so that it's still accessible
				awsServiceHome := *awsService
				awsServiceHome.Id = "home"
				awsService.SubServices = append(
					[]core.AwsService{
						awsServiceHome,
					},
					awsService.SubServices...,
				)

				if len(awsService.SubServices) > 1 {
					splitQuery = strings.Split(query, " ")
					if len(splitQuery) > 1 {
						subServiceId := splitQuery[0]
						var subService *core.AwsService
						for i := range awsService.SubServices {
							if awsService.SubServices[i].Id == subServiceId {
								subService = &awsService.SubServices[i]
								break
							}
						}
						if subService != nil {
							query = strings.Join(splitQuery[1:], " ")
							id = id + "_" + subServiceId
							log.Println("id", id)
							populater := PopulatersByServiceId[id]
							if populater != nil {
								log.Printf("using populater associated with %s", id)
								err := populater(wf, query, session, forceFetch, fullQuery)
								if err != nil {
									wf.FatalError(err)
								}
								return query
							}
						}
					}
				}
				log.Printf("filtering on subServices for %s", id)
				query = strings.TrimSpace(strings.Join(splitQuery, " "))
				PopulateSubServices(wf, *awsService)
				return query
			}
		}
	}

	PopulateServices(wf, awsServices)
	return query
}