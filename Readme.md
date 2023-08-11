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
#### Build
- `git clone https://github.com/akopulko/ffiiitc.git`
- `docker build -t ffiiitc:latest .`
#### Run
- Stop `docker compose -f docker-compose.yml down`
- Modify your FireFly III docker compose file add the following
```yaml
  fftc:
    image: ffiiitc:latest
    hostname: fftc
    networks:
      - firefly_iii
    restart: always
    container_name: ffiiitc
    environment:
      - FF_API_KEY=<YOUR_PAT_GOES_HERE>
      - FF_APP_URL=http://app:8080
    volumes:
      - ffiiitc-data:/app/data
    ports:
     - '8082:8080'
    depends_on:
     - app
volumes:
    ...
   ffiiitc-data:
```
- Start `docker compose -f docker-compose.yml up -d`
#### Configure Web Hooks in FireFly
In `FireFly` go to `Automation -> Webhooks` and click `Create new webhook`
- Create webhook for transaction classification
```yaml
title: classify
trigger: after transaction creation
response: transaction details
delivery: json
url: http://fftc:8080
active: checked
```


### Troubleshooting
You can check `ffiiitc` logs to see if there are any errors:<br> `docker compose logs fftc -f`