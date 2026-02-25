# resource "wallarm_rule_disable_stamp" "disable_stamp_test" {
#   point = [data.wallarm_hits.hits_test-2.hits[0].point]
#   stamp = data.wallarm_hits.hits_test-2.hits[0].stamps[0]
  
#   dynamic "action" {
#     for_each = data.wallarm_hits.hits_test-2.hits[0].action

#     content {
#       point = action.value.point
#       value = action.value.value
#       type = action.value.type == "" ? null : action.value.type
#     }
#   }

# }
