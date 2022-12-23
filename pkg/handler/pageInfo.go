package handler

const PAGESIZEALL = 0

type pageInfo struct {
	pageSize     int
	currentTotal int

	currentOffset int

	totalPage   int
	currentPage int

	totalCount int
}

func (p *pageInfo) updatePageInfo(pageSize, currTotal, offset, total int) {
	p.pageSize = pageSize
	p.currentTotal = currTotal
	p.currentOffset = offset
	p.totalCount = total
	p.update()
}

func (p *pageInfo) update() {
	// 根据 pageSize和total值 更新  totalPage currentPage
	if p.pageSize <= 0 {
		p.totalPage = 1
		p.currentPage = 1
		return
	}
	pageSize := p.pageSize
	totalCount := p.totalCount

	switch totalCount % pageSize {
	case 0:
		p.totalPage = totalCount / pageSize
	default:
		p.totalPage = (totalCount / pageSize) + 1
	}
	switch p.currentOffset % pageSize {
	case 0:
		p.currentPage = p.currentOffset / pageSize
	default:
		p.currentPage = (p.currentOffset / pageSize) + 1
	}
}

func (p *pageInfo) CurrentTotalCount() int {
	return p.currentTotal
}

func (p *pageInfo) TotalPage() int {
	return p.totalPage
}

func (p *pageInfo) TotalCount() int {
	return p.totalCount
}

func (p *pageInfo) PageSize() int {
	return p.pageSize
}

func (p *pageInfo) CurrentPage() int {
	return p.currentPage
}

func (p *pageInfo) CurrentOffSet() int {
	return p.currentOffset
}
