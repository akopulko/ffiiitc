# FireFly III Transaction Classification

This little web service performs transaction classification and integrates with [FireFly III](https://github.com/firefly-iii/firefly-iii) (A free and open source personal finance manager) via web hooks.

### What is does?

Every time you add new transaction to FireFly III, either manually or via import tool, web hook will trigger and provide transaction description to `ffiiitc`. It will then be classified using [Naive Bayesian Classification](https://en.wikipedia.org/wiki/Naive_Bayes_classifier) and transaction will be updated with matching category.

> Naive Bayesian classifier go package used by `ffiiitc` is available [here](https://github.com/navossoc/bayesian). Please read the [license](https://github.com/navossoc/bayesian/blob/master/LICENSE).

### How to run?

#### Pre-requisites

- [Docker desktop](https://www.docker.com/products/docker-desktop/) or any other form of running containers on your computer
- [FireFly III](https://github.com/firefly-iii/firefly-iii) up and running as per [docs](https://docs.firefly-iii.org/firefly-iii/installation/docker/?mtm_campaign=docu-internal&mtm_kwd=docker)
- At least **one or two statements** imported into FireFly with transactions manually **categorised**. This is required for classifier to train on your dataset and is very important.
- Have personal access token (PAT) generated in FireFly III. Go to `Options->Profile->OAuth` click `Create new token`

##### Docker Compose

- `git clone https://github.com/akopulko/ffiiitc.git`
- `docker buildx build --load --platform=linux/amd64 -t ffiiitc:latest .`

#### Run

##### Docker Compose

- Stop `docker compose -f docker-compose.yml down`
- Modify your FireFly III docker compose file add the following

```yaml
  fftc:
    image: akopulko/ffiiitc:latest
    hostname: fftc
    networks:
      - firefly_iii
    restart: always
    container_name: ffiiitc
    environment:
      - FF_API_KEY=<YOUR_PAT_GOES_HERE>
      - FF_APP_URL=<FIREFLY_ADDRESS:PORT>
    volumes:
      - ffiiitc-data:/app/data
    ports:
     - '<EXPOSED_PORT>:8080'
    depends_on:
     - app
volumes:
    ...
   ffiiitc-data:
```

- Start `docker compose -f docker-compose.yml up -d`

#### Docker

```
docker run
  -d
  --name='ffiiitc'
  -e 'FF_API_KEY'='<YOUR_PAT_GOES_HERE>'
  -e 'FF_APP_URL'='<FIREFLY_ADDRESS:PORT>'
  -p '<EXPOSED_PORT>:8080'
  -v '<TRAINED_MODEL_FOLDER>':'/app/data':'rw' 'ffiiitc'
```

#### Configure Web Hooks in FireFly

In `FireFly` go to `Automation -> Webhooks` and click `Create new webhook`

- Create webhook for transaction classification

```yaml
title: classify
trigger: after transaction creation
response: transaction details
delivery: json
url: http://fftc:<EXPOSED_PORT>/classify
active: checked
```

### Troubleshooting

#### Logs
You can check `ffiiitc` logs to see if there are any errors:<br> `docker compose logs fftc -f`

#### Forced training of your model
There is also option available to force train the model from your transactions if required. 
To trigger force train run the following command and restart `fftc` container:
`curl -i http://localhost:<EXPOSED_PORT>/train` where `EXPOSED_PORT` is the port you provided in your docker compose for `fftc`. 
As always, you can check logs to see if model was successfully regenerated. 