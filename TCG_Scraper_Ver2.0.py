import sys
import gspread
import time
import requests
from bs4 import BeautifulSoup
from oauth2client.service_account import ServiceAccountCredentials


# "client_email": "tcgscraper@tcg-scraper.iam.gserviceaccount.com"
# TODO: Optimize Speed
# TODO: Work around JAVA for TCG Player

def main():
    print("Starting Scraper")

    headers = {'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:76.0) Gecko/20100101 Firefox/76.0'}
    scope = ['https://spreadsheets.google.com/feeds', 'https://www.googleapis.com/auth/drive']
    creds = ServiceAccountCredentials.from_json_keyfile_name('tcgcreds.json', scope)
    client = gspread.authorize(creds)
    spreadsheet = client.open('Square TCG Scraper Ver1.1')
    sheet = spreadsheet.get_worksheet(1)


    rows = sheet.get_all_records()
    #Get all of the URLs in one object
    da_url_list = sheet.batch_get(['Y2:Y'+str(len(rows) + 1)])[0]
    tcg_url_list = sheet.batch_get(['Z2:Z' + str(len(rows) + 1)])[0]
    tnt_url_list = sheet.batch_get(['AA2:AA' + str(len(rows) + 1)])[0]
    cs_url_list = sheet.batch_get(['AB2:AB' + str(len(rows) + 1)])[0]

    row = 2
    for start_index in range(0, len(da_url_list), 10):

        da_batch_list = []


        for url in da_url_list[start_index:start_index+10]: #The first item is a list containing all the information
            da_range = "A" + str(row) + ":F" + str(row)

            if url:
                print("Reading url info for ", "line:",row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")

                da_title = soup.find('h1').get_text().strip()
                try:
                    #Grab price info
                    da_price = float(soup.find(class_='price large').get_text().strip("$").replace(',', ''))
                except AttributeError:
                    da_price = ""

                try:
                    #Need to get around no eight columns 'NoneType' Object
                    da_desc = soup.find(class_='eight columns').get_text().strip() #grab description
                except AttributeError:
                    da_desc = soup.find(id='moredetailsTab').get_text().strip()

                da_pic = soup.find(class_='panel radius').findChild()['href'] #grab pic info
                da_upc = "UPC/Barcode" in soup.find(id='itemdetailsTab').find_all("li")[3].get_text()
                da_upc_text = ""
                da_stock_text = ""

                #logic for barcode information
                if da_upc:
                    da_upc_text = soup.find(id='itemdetailsTab').find_all("li")[3].get_text().strip("UPC/Barcode:")

                #logic for stock information
                if da_price == "":
                    da_stock_text = "Out Of Stock"
                else:
                    da_stock_text = "In Stock"


                da_batch_list.append({'range': da_range, 'values': [[da_title, da_price, da_desc, da_pic, da_upc_text, da_stock_text]]})
            row += 1
        da_total_list = sheet.batch_update(da_batch_list)

    row = 2
    for start_index in range(0, len(tcg_url_list), 10):

        tcg_batch_list = []

        for url in tcg_url_list[start_index:start_index + 10]:  # The first item is a list containing all the information
            tcg_range = "G" + str(row) + ":K" + str(row)

            if url:
                print("Reading url info for ", "line:",row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")
                print(soup)
                tcg_title = soup.find("div", {"class": 'product-details__name'})  #Grabs item Name
                print(tcg_title)
                try:
                    #Grabs price
                    tcg_price = soup.find(class_='spotlight__price')
                except AttributeError:
                    tcg_price = ""

                tcg_desc = soup.find(class_='product-details__details-description').get_text() #Grabs Description
                tcg_pic = soup.find(class_='product-details__info').findChild('img')['src']
                tcg_stock_text = ""

                # logic for stock information
                if tcg_price == "":
                    tcg_stock_text = "Out Of Stock"
                else:
                    tcg_stock_text = "In Stock"

                tcg_batch_list.append({'range': tcg_range, 'values': [[tcg_title, tcg_price, tcg_desc, tcg_pic, tcg_stock_text]]})
            row += 1
        tcg_total_list = sheet.batch_update(tcg_batch_list)


    row = 2
    for start_index in range(0, len(tnt_url_list), 10):

        tnt_batch_list = []

        for url in tnt_url_list[start_index:start_index + 10]:  # The first item is a list containing all the information
            tnt_range = "L" + str(row) + ":O" + str(row)

            if url:
                print("Reading url info for ", "line:",row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")

                tnt_title = soup.find('h1').get_text().strip()
                tnt_pics = soup.find(id='main-prod-img').findChild()['data-src']

                try:
                    tnt_price = float(soup.find(class_="d-flex flex-column").span.get_text().strip('$'))
                except AttributeError:
                    tnt_price = ""

                tnt_stock_text = ""

                # logic for stock information
                if tnt_price == "":
                    tnt_stock_text = "Out of Stock"
                else:
                    tnt_stock_text = "In Stock"

                tnt_batch_list.append({'range': tnt_range, 'values': [[tnt_title, tnt_price, tnt_pics, tnt_stock_text]]})
            row += 1
        tnt_total_list = sheet.batch_update(tnt_batch_list)

    #Collector Store Note: Sometimes pages go 404 and the scraper will stop
    #Delete the 404 url and the scraper will resume
    row = 2
    for start_index in range(0, len(cs_url_list), 10):

        cs_batch_list = []

        for url in cs_url_list[start_index:start_index + 10]:  # The first item is a list containing all the information
            cs_range = "P" + str(row) + ":S" + str(row)

            if url:
                print("Reading url info for ", "line:", row, url[0])

                try:
                    response = requests.get(url[0], headers)
                except requests.exceptions.MissingSchema:
                    return

                soup = BeautifulSoup(response.content, features="lxml")

                cs_title = soup.find('h1').get_text().strip()

                try:
                    cs_price = float(soup.find(class_="price price--withoutTax").get_text().strip("$"))
                except AttributeError:
                    cs_price = ""

                cs_pic = soup.find(class_='productView-img-container').findChild()['href']

                # logic for stock information
                if cs_price == "":
                    cs_stock_text = "Out of Stock"
                else:
                    cs_stock_text = "In Stock"

                cs_batch_list.append({'range': cs_range, 'values': [[cs_title, cs_price, cs_pic, cs_stock_text]]})
            row += 1
        cs_total_list = sheet.batch_update(cs_batch_list)

def
if __name__ == "__main__":
    main()
