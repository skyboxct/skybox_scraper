import gspread
import pyppeteer
from bs4 import BeautifulSoup
from oauth2client.service_account import ServiceAccountCredentials
from requests_html import HTMLSession

MAX_RETRIES = 5

da_output_range = ["A", "F"]
sc_output_range = ["G", "J"]


def main():
    print("Starting Scraper")

    scope = ['https://spreadsheets.google.com/feeds', 'https://www.googleapis.com/auth/drive']
    creds = ServiceAccountCredentials.from_json_keyfile_name('sportscreds.json', scope)
    client = gspread.authorize(creds)
    spreadsheet = client.open('Shopify Sports Scraper Ver2.0')
    sheet = spreadsheet.get_worksheet(1)
    records = sheet.get_all_records()

    scrape_products(records, sheet)


def scrape_products(records, sheet):
    # Dictionary to hold host configurations
    host_configs = {
        "DA": {"range": da_output_range, "parser_func": parse_da},
        "SC": {"range": sc_output_range, "parser_func": parse_sc}
    }

    row = 2
    for start_index in range(0, len(records), 10):
        batch_list = []
        for product in records[start_index:start_index + 10]:
            print("\nReading product information for row", row)
            urls = [product["DA url"], product["SC url"]]
            product_pages = get_product_pages(urls)
            for page in product_pages:
                print("   Reading page for", page["url"])
                host = page["host"]
                page_html = page["html"]

                if host == "TCG":
                    parser_parameter = page_html
                else:
                    parser_parameter = BeautifulSoup(page_html.html.raw_html, features="lxml")

                output_range = host_configs[host]["range"][0] + str(row) + ":" + host_configs[host]["range"][1] + str(
                    row)
                values = retry(host_configs[host]["parser_func"], parser_parameter, MAX_RETRIES)

                if output_range and values:
                    print("   Values: ", values)
                    batch_list.append({'range': output_range, 'values': values})
            row += 1
        print("\nPushing lines", start_index + 2, "-", start_index + 12, "to sheet\n")
        sheet.batch_update(batch_list)
    # TODO: roll up parsing errors into output before exit


def parse_da(soup):
    title = soup.find('h1').get_text().strip()

    try:
        # Grab Sale price info
        price = float(soup.find(class_='price discount large').get_text().strip("$").replace(',', ''))
        print("DISCOUNTED! REAL PRICE: ", price)
    except AttributeError:
        price = ""

    if price == "":
        try:
            # Grab price info
            price = float(soup.find(class_='price large').get_text().strip("$").replace(',', ''))
        except AttributeError:
            price = ""

    desc = soup.find(class_='eight columns').get_text().strip()  # grab description
    pic = soup.find(class_='panel radius').findChild()['href']  # grab pic info
    upc = "UPC/Barcode" in soup.find(id='itemdetailsTab').find_all("li")[3].get_text()
    upc_text = ""
    stock_text = ""

    # logic for barcode information
    if upc:
        upc_text = soup.find(id='itemdetailsTab').find_all("li")[3].get_text().strip("UPC/Barcode:")

    # logic for stock information
    if price == "":
        stock_text = "Out Of Stock"
    else:
        stock_text = "In Stock"

    return [[title, price, desc, pic, upc_text, stock_text]]


def parse_sc(soup):
    title = soup.find(itemprop='name').get_text().strip()
    try:
        price = float(soup.find(itemprop='price').get_text().strip().replace(',', ''))
    except AttributeError:
        price = ""

    pic = soup.find('source').find_next()['srcset']
    sc_stock = "Sorry, but this item is currently out of stock." in soup.find(
        class_="five columns p-buy-box").get_text()
    stock_text = ""

    # logic for stock information
    if sc_stock:
        price = ""
        stock_text = "Out of Stock"
    else:
        stock_text = "In Stock"

    return [[title, price, pic, stock_text]]


def get_product_pages(urls):
    page_dict = []
    for url in urls:
        if url:
            session = HTMLSession()
            session.headers = {
                'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:76.0) Gecko/20100101 Firefox/76.0'}
            host = get_host(url)
            print("Connecting to", url)
            response_html = retry(session.get, url, MAX_RETRIES)
            if response_html:
                try:
                    print("Rendering html")
                    response_html.html.render(timeout=100, retries=MAX_RETRIES)
                    page_dict.append({"host": host, "url": url, "html": response_html})
                except pyppeteer.errors.TimeoutError as ex:
                    print("Failed to render page for", url, "after", MAX_RETRIES, "attempts:", type(ex).__name__)
                except pyppeteer.errors.PageError as ex:
                    print("Failed to render page for", url, "after", MAX_RETRIES, "attempts:", type(ex).__name__)
            session.close()
    return page_dict


def get_host(url):
    if "dacardworld.com" in url:
        return "DA"
    elif "steelcitycollectibles.com" in url:
        return "SC"
    else:
        print("Unknown host for url:", url)
        return "Unknown"


def retry(func, param, max_retries):
    for i in range(0, max_retries + 1):
        try:
            results = func(param)
        except Exception as ex:
            template = "An exception of type {0} occurred."
            message = template.format(type(ex).__name__)
        else:
            return results
    print("Function '", func.__name__, "' failed after", i + 1, "attempts:", message)
    return None


if __name__ == "__main__":
    main()

#   "client_email": "amazon-scrape@amazon-data-scrape.iam.gserviceaccount.com",

