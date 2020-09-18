# Crossposter

Application for forwarding posts between different services. The producer-consumer pattern with multiple topics for cross-posting.

## Support

| service | producer | consumer | web endpoint |
|:--|:-:|:-:|:-:|
| Instagram | x | x | |
| Pikabu | x | | |
| Reddit | x | | |
| RSS | x | x | x |
| Telegram | x | x | |
| test | x | x | |
| Twitter | x | x | |
| Vkontakte | x | x | |

## Config

YAML file:

```yaml
---
producers:
  - type: instagram
    sources:
    - account_name
    options:
      user: <...>
      password: <...>
    topics:
    - topic_for_producing
  - type: pikabu
    sources:
    - community/name
    - tag/name
    - any/location/with/posts
    topics:
    - topic_for_producing
  - type: reddit
    sources:
    - subreddit_name
    topics:
    - topic_for_producing
  - type: rss
    sources:
    - http://domain.com/rss.xml
    topics:
    - topic_for_producing
  - type: telegram
    options:
      token: <...>
    sources:
    - channel_name
    - -1000000000000  # channel ID
    topics:
    - topic_for_producing
  - type: test
    description: For test of producing
    sources:
    - test
    options:
      title: TEST
      url: http://test.com
      author: Tester
      text: <b>Test</b> message
      attachment: http://test.com/image.jpg
    topics:
    - topic_for_producing
  - type: twitter
    options:
      key: <...>
      key_secret: <...>
      token: <...>
      token_secret: <...>
    sources:
    - screen_name
    topics:
    - topic_for_producing
  - type: vk
    options:
      token: <...>
      # Or
      user: <...>
      password: <...>
    sources:
    - group_name
    topics:
    - topic_for_producing
consumers:
  - type: instagram
    options:
      user: <...>
      password: <...>
    topics:
    - topic_for_consuming
  - type: rss
    description: Site news feed
    options:  # For generated RSS feed
      title: RSS feed
      link: http://domain.com/
    destination:
    - news  # location for web service: localhost/rss/news
    topics:
    - topic_for_consuming
  - type: telegram
    options:
      token: <...>
    topics:
    - topic_for_consuming
    destinations:
    - channel_name
    - -1000000000000  # channel ID
    topics:
    - topic_for_consuming
  - type: test
    description: Print post in log
    topics:
    - topic_for_consuming
  - type: twitter
    options:
      key: <...>
      key_secret: <...>
      token: <...>
      token_secret: <...>
    topics:
    - topic_for_consuming
  - type: vk
    options:
      token: <...>
      # Or
      user: <...>
      password: <...>
    destinations:
    - public_name
    topics:
    - topic_for_consuming
```
